# Gnomock InfluxDB

Gnomock InfluxDB is a [Gnomock](https://github.com/orlangure/gnomock) preset
for running tests against a real InfluxDB container, without mocks.

```go
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
)

func TestInfluxDB(t *testing.T) {
	username := "gnomock"
	password := "gnomock-password"
	org := "gnomock-org"
	bucket := "gnomock-bucket"
	token := "gnomock-influxdb-token"

	p := influxdb.Preset(
		influxdb.WithVersion("alpine"),
		influxdb.WithAuthToken(token),
		influxdb.WithBucket(bucket),
		influxdb.WithOrg(org),
		influxdb.WithUser(username, password),
	)
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
}
```

