# Gnomock Memcached

Gnomock Memcached is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real Memcached container, without mocks.

```go
package memcached_test

import (
	"fmt"
	"testing"

	memcachedclient "github.com/bradfitz/gomemcache/memcache"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/memcached"
)

func ExamplePreset() {
	vs := make(map[string][]byte)

	// String
	vs["a"] = []byte("foo")
	// Number
	vsb := make([]byte, 4)
	binary.LittleEndian.PutUint32(vsb, 42)
	vs["b"] = vsb

	p := memcached.Preset(memcached.WithValues(vs))
	container, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	addr := container.DefaultAddress()
	client := memcachedclient.New(addr)

	itemA, _ := client.Get("a")
	fmt.Println(string(itemA.Value))

	itemB, _ := client.Get("b")
	fmt.Println(binary.LittleEndian.Uint32(itemB.Value))

	// Output:
	// foo
	// 42
}
```
