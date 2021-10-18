package gnomockd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/memcached"
	"github.com/stretchr/testify/require"
)

func TestMemcached(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/memcached.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/memcached", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	client := memcache.New(c.DefaultAddress())

	itemA, err := client.Get("answer")
	require.NoError(t, err)
	valueA, err := strconv.ParseInt(string(itemA.Value), 10, 32)
	require.NoError(t, err)
	require.Equal(t, 42, int(valueA))

	itemB, err := client.Get("bar")
	require.NoError(t, err)
	require.Equal(t, "foo", string(itemB.Value))

	itemC, err := client.Get("computer")
	require.NoError(t, err)
	require.Equal(t, "hal9000", string(itemC.Value))

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
