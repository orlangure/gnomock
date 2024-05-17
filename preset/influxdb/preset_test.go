package influxdb_test

import (
	"context"
	"fmt"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/influxdb"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	for _, version := range []string{"alpine"} {
		t.Run("version", func(t *testing.T) {
			t.Run("with default values", testPreset(version, true))
			t.Run("with custom values", testPreset(version, false))
		})
	}

	t.Run("with default version", testPreset("", true))
}

func testPreset(version string, useDefaults bool) func(t *testing.T) {
	// default values
	username := "gnomock"
	password := "gnomock-password"
	org := "gnomock-org"
	bucket := "gnomock-bucket"
	token := "gnomock-influxdb-token"

	if !useDefaults {
		username = "foobar"
		password = "password"
		org = "my-org"
		bucket = "some-bucket"
		token = "foobar-token"
	}

	opts := []influxdb.Option{influxdb.WithVersion(version)}
	if !useDefaults {
		opts = append(
			opts,
			influxdb.WithAuthToken(token),
			influxdb.WithBucket(bucket),
			influxdb.WithOrg(org),
			influxdb.WithUser(username, password),
		)
	}

	return func(t *testing.T) {
		p := influxdb.Preset(opts...)
		container, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		addr := fmt.Sprintf("http://%s", container.DefaultAddress())
		client := influxdb2.NewClient(addr, token)
		ctx := context.Background()

		h, err := client.Health(ctx)
		require.NoError(t, err)
		require.Equal(t, domain.HealthCheckStatusPass, h.Status)

		buckets, err := client.BucketsAPI().GetBuckets(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, buckets)

		bucketNames := make([]string, 0, len(*buckets))
		for _, bucket := range *buckets {
			bucketNames = append(bucketNames, bucket.Name)
		}

		require.Contains(t, bucketNames, bucket)

		users, err := client.UsersAPI().GetUsers(ctx)
		require.NoError(t, err)

		userNames := make([]string, 0, len(*users))
		for _, user := range *users {
			userNames = append(userNames, user.Name)
		}

		require.Contains(t, userNames, username)

		orgs, err := client.OrganizationsAPI().GetOrganizations(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, orgs)

		orgNames := make([]string, 0, len(*orgs))
		for _, org := range *orgs {
			orgNames = append(orgNames, org.Name)
		}

		require.Contains(t, orgNames, org)
		client.Close()
	}
}
