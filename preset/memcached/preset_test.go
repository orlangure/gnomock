// +build !nopreset

package memcached_test

import (
	"encoding/binary"
	"testing"

	memcachedclient "github.com/bradfitz/gomemcache/memcache"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/memcached"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	vs := make(map[string][]byte)

	// String
	vs["a"] = []byte("foo")
	// Number
	vsb := make([]byte, 4)
	binary.LittleEndian.PutUint32(vsb, 42)
	vs["b"] = vsb

	p := memcached.Preset(memcached.WithValues(vs))
	container, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	require.NoError(t, err)

	addr := container.DefaultAddress()
	client := memcachedclient.New(addr)

	itemA, err := client.Get("a")
	require.NoError(t, err)
	require.Equal(t, "foo", string(itemA.Value))

	itemB, err := client.Get("b")
	require.NoError(t, err)
	require.Equal(t, uint32(42), binary.LittleEndian.Uint32(itemB.Value))
}
