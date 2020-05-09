# Gnomock Redis ![Build](https://github.com/orlangure/gnomock-redis/workflows/Build/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock-redis)](https://goreportcard.com/report/github.com/orlangure/gnomock-redis)

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
	container, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	addr := container.Address(gnomock.DefaultPort)
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
```
