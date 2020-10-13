package redis_test

import (
	"testing"

	redisclient "github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/redis"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	vs := make(map[string]interface{})

	vs["a"] = "foo"
	vs["b"] = 42
	vs["c"] = true

	p := redis.Preset(redis.WithValues(vs))
	container, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	require.NoError(t, err)

	addr := container.DefaultAddress()
	client := redisclient.NewClient(&redisclient.Options{Addr: addr})

	var str string

	require.NoError(t, client.Get("a").Scan(&str))
	require.Equal(t, "foo", str)

	var number int

	require.NoError(t, client.Get("b").Scan(&number))
	require.Equal(t, 42, number)

	var flag bool

	require.NoError(t, client.Get("c").Scan(&flag))
	require.True(t, flag)
}

func TestRedis_wrongValue(t *testing.T) {
	t.Parallel()

	vs := map[string]interface{}{"a": []string{"b", "c"}}
	p := redis.Preset(redis.WithValues(vs))

	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.Error(t, err)
}
