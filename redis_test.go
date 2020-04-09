package redis_test

import (
	"fmt"
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
	mockredis "github.com/orlangure/gnomock-redis"
	"github.com/stretchr/testify/require"
)

func ExampleRedis() {
	vs := make(map[string]interface{})

	vs["a"] = "foo"
	vs["b"] = 42
	vs["c"] = true

	p := mockredis.Preset(mockredis.WithValues(vs))
	container, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	addr := container.DefaultAddress()
	client := redis.NewClient(&redis.Options{Addr: addr})

	fmt.Println(client.Get("a").Result())

	var number int

	err := client.Get("b").Scan(&number)
	fmt.Println(number, err)

	var flag bool

	err = client.Get("c").Scan(&flag)
	fmt.Println(flag, err)

	// Output:
	// foo <nil>
	// 42 <nil>
	// true <nil>
}

func TestRedis_wrongValue(t *testing.T) {
	vs := map[string]interface{}{"a": []string{"b", "c"}}
	p := mockredis.Preset(mockredis.WithValues(vs))

	c, err := gnomock.Start(p)

	defer func() {
		err := gnomock.Stop(c)
		require.NoError(t, err)
	}()

	require.Error(t, err)
}
