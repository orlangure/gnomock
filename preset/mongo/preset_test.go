package mongo_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/mongo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"4.4", "3.6.21", "5.0"} {
		t.Run(version, testPreset(version))
	}
}

func testPreset(version string) func(t *testing.T) {
	return func(t *testing.T) {
		p := mongo.Preset(
			mongo.WithData("./testdata/"),
			mongo.WithUser("gnomock", "gnomick"),
			mongo.WithVersion(version),
		)
		c, err := gnomock.Start(p, gnomock.WithLogWriter(os.Stdout))

		defer func() { require.NoError(t, gnomock.Stop(c)) }()

		require.NoError(t, err)

		ctx := context.Background()
		addr := c.DefaultAddress()
		uri := fmt.Sprintf("mongodb://%s:%s@%s", "gnomock", "gnomick", addr)
		clientOptions := mongooptions.Client().ApplyURI(uri)

		client, err := mongodb.Connect(ctx, clientOptions)
		require.NoError(t, err)

		// see testdata folder to verify names/numbers
		count, err := client.Database("db1").Collection("users").CountDocuments(ctx, bson.D{})
		require.NoError(t, err)
		require.Equal(t, int64(10), count)

		count, err = client.Database("db2").Collection("customers").CountDocuments(ctx, bson.D{})
		require.NoError(t, err)
		require.Equal(t, int64(5), count)

		count, err = client.Database("db2").Collection("countries").CountDocuments(ctx, bson.D{})
		require.NoError(t, err)
		require.Equal(t, int64(3), count)
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := mongo.Preset()
	c, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	ctx := context.Background()
	addr := c.DefaultAddress()
	uri := fmt.Sprintf("mongodb://%s:%s@%s", "gnomock", "gnomick", addr)
	clientOptions := mongooptions.Client().ApplyURI(uri)

	client, err := mongodb.Connect(ctx, clientOptions)
	require.NoError(t, err)
	require.NoError(t, client.Disconnect(ctx))
}

func TestPreset_wrongDataFolder(t *testing.T) {
	t.Parallel()

	p := mongo.Preset(mongo.WithData("./bad-path"))
	c, err := gnomock.Start(p)
	require.Error(t, err)
	require.NoError(t, gnomock.Stop(c))
}
