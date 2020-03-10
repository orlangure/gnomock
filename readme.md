# Gnomock Redis

Gnomock Redis is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real Redis container, without mocks.

```go
package redis_test

import (
	"fmt"
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
	mockredis "github.com/orlangure/gnomock-redis"
)

func ExampleRedis() {
	vs := make(map[string]interface{})

	vs["a"] = "foo"
	vs["b"] = 42
	vs["c"] = true

	p := mockredis.Preset(mockredis.WithValues(vs))

	c, err := gnomock.StartPreset(p)

	defer func() {
		err := gnomock.Stop(c)
		if err != nil {
			panic(err)
		}
	}()

	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{Addr: c.Address()})

	_, err = client.Ping().Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(client.Get("a").Result())

	var number int

	err = client.Get("b").Scan(&number)
	fmt.Println(number, err)

	var flag bool

	err = client.Get("c").Scan(&flag)
	fmt.Println(flag, err)

	// Output:
	// foo <nil>
	// 42 <nil>
	// true <nil>
}
```
