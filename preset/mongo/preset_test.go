package mongo_test

import (
	"context"
	"fmt"
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

	p := mongo.Preset(
		mongo.WithData("./testdata/"),
		mongo.WithUser("gnomock", "gnomick"),
		mongo.WithVersion("4"),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	addr := c.DefaultAddress()
	uri := fmt.Sprintf("mongodb://%s:%s@%s", "gnomock", "gnomick", addr)
	clientOptions := mongooptions.Client().ApplyURI(uri)

	client, err := mongodb.NewClient(clientOptions)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
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

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := mongo.Preset()
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	addr := c.DefaultAddress()
	uri := fmt.Sprintf("mongodb://%s:%s@%s", "gnomock", "gnomick", addr)
	clientOptions := mongooptions.Client().ApplyURI(uri)

	client, err := mongodb.NewClient(clientOptions)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, client.Disconnect(ctx))
}
