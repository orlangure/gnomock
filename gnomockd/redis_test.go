// +build !noserver

package gnomockd_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/gnomockd"
	"github.com/stretchr/testify/require"
)

//nolint:bodyclose
func TestRedis(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := ioutil.ReadFile("./testdata/redis.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/redis", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{Addr: c.DefaultAddress()})

	answer := 0

	require.NoError(t, client.Get("answer").Scan(&answer))
	require.Equal(t, 42, answer)

	var isGood bool

	require.NoError(t, client.Get("is_good").Scan(&isGood))
	require.True(t, isGood)

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
