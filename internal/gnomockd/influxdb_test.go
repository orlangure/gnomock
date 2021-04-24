package gnomockd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/influxdb"
	"github.com/stretchr/testify/require"
)

func TestInfluxDB(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := ioutil.ReadFile("./testdata/influxdb.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/influxdb", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)
	require.NotEmpty(t, c.DefaultAddress())

	addr := fmt.Sprintf("http://%s", c.DefaultAddress())
	client := influxdb2.NewClient(addr, "real-auth-token")
	ctx := context.Background()

	health, err := client.Health(ctx)
	require.NoError(t, err)
	require.Equal(t, domain.HealthCheckStatusPass, health.Status)

	buckets, err := client.BucketsAPI().GetBuckets(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, buckets)

	bucketNames := make([]string, 0, len(*buckets))
	for _, bucket := range *buckets {
		bucketNames = append(bucketNames, bucket.Name)
	}

	require.Contains(t, bucketNames, "a-bucket")

	users, err := client.UsersAPI().GetUsers(ctx)
	require.NoError(t, err)

	userNames := make([]string, 0, len(*users))
	for _, user := range *users {
		userNames = append(userNames, user.Name)
	}

	require.Contains(t, userNames, "gnomock")

	orgs, err := client.OrganizationsAPI().GetOrganizations(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, orgs)

	orgNames := make([]string, 0, len(*orgs))
	for _, org := range *orgs {
		orgNames = append(orgNames, org.Name)
	}

	require.Contains(t, orgNames, "gnomorg")

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
