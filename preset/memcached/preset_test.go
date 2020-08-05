// +build !nopreset

package memcached_test

import (
	"encoding/binary"
	"strconv"
	"testing"

	memcachedclient "github.com/bradfitz/gomemcache/memcache"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/memcached"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	vs1 := make(map[string][]byte)

	// String
	vs1["a"] = []byte("foo")
	// Number as byte slice
	vs1b := make([]byte, 4)
	binary.LittleEndian.PutUint32(vs1b, 42)
	vs1["b"] = vs1b
	// Number as string
	vs1["c"] = []byte(strconv.FormatInt(42, 10))

	vs2 := make(map[string]string)

	vs2["d"] = "foo"

	p := memcached.Preset(memcached.WithByteValues(vs1), memcached.WithValues(vs2))
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
	require.Equal(t, 42, int(binary.LittleEndian.Uint32(itemB.Value)))

	itemC, err := client.Get("c")
	require.NoError(t, err)
	valueC, err := strconv.ParseInt(string(itemC.Value), 10, 32)
	require.NoError(t, err)
	require.Equal(t, 42, int(valueC))

	itemD, err := client.Get("d")
	require.NoError(t, err)
	require.Equal(t, "foo", string(itemD.Value))
}
