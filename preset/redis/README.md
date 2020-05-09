# Gnomock Redis

Gnomock Redis is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real Redis container, without mocks.

```go
package redis_test

import (
	"fmt"
	"testing"

	redisclient "github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/redis"
)

func ExamplePreset() {
	vs := make(map[string]interface{})

	vs["a"] = "foo"
	vs["b"] = 42
	vs["c"] = true

	p := redis.Preset(redis.WithValues(vs))
	container, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	addr := container.DefaultAddress()
	client := redisclient.NewClient(&redisclient.Options{Addr: addr})

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
