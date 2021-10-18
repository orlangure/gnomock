package gnomockd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	"github.com/orlangure/gnomock/internal/israce"
	_ "github.com/orlangure/gnomock/preset/elastic"
	"github.com/stretchr/testify/require"
)

func TestElastic(t *testing.T) {
	if israce.Enabled {
		t.Skip("elastic tests can't run with race detector due to https://github.com/elastic/go-elasticsearch/issues/147")
	}

	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/elastic.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/elastic", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	require.NoError(t, json.Unmarshal(body, &c))

	cfg := elasticsearch.Config{
		Addresses:    []string{fmt.Sprintf("http://%s", c.DefaultAddress())},
		DisableRetry: true,
	}

	client, err := elasticsearch.NewClient(cfg)
	require.NoError(t, err)

	searchResult, err := client.Search(
		client.Search.WithIndex("titles"),
		client.Search.WithQuery("gnomock"),
	)
	require.NoError(t, err)
	require.False(t, searchResult.IsError(), searchResult.String())

	var out struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}

	require.NoError(t, json.NewDecoder(searchResult.Body).Decode(&out))
	require.NoError(t, searchResult.Body.Close())
	require.Equal(t, 1, out.Hits.Total.Value)

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
