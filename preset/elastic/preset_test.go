package elastic_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/israce"
	"github.com/orlangure/gnomock/preset/elastic"
	"github.com/stretchr/testify/require"
)

// these tests have trouble running in parallel, probably due to limited
// resources

func TestPreset(t *testing.T) {
	if israce.Enabled {
		t.Skip("elastic tests can't run with race detector due to https://github.com/elastic/go-elasticsearch/issues/147")
	}

	for _, version := range []string{"8.7.0", "7.9.3", "6.8.13", "5.6"} {
		t.Run(version, testPreset(version))
	}
}

func testPreset(version string) func(t *testing.T) {
	return func(t *testing.T) {
		p := elastic.Preset(
			elastic.WithVersion("7.9.2"),
			elastic.WithInputFile("./testdata/titles"),
			elastic.WithInputFile("./testdata/names"),
		)

		c, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(c)) }()

		cfg := elasticsearch.Config{
			Addresses:    []string{fmt.Sprintf("http://%s", c.DefaultAddress())},
			DisableRetry: true,
		}

		client, err := elasticsearch.NewClient(cfg)
		require.NoError(t, err)

		res, err := client.Search(
			client.Search.WithIndex("titles"),
			client.Search.WithQuery("gnomock"),
		)
		require.NoError(t, err)
		require.False(t, res.IsError(), res.String())

		var out struct {
			Hits struct {
				Total struct {
					Value int `json:"value"`
				} `json:"total"`
			} `json:"hits"`
		}

		require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
		require.NoError(t, res.Body.Close())
		require.Equal(t, 1, out.Hits.Total.Value)

		res, err = client.Search(
			client.Search.WithIndex("titles"),
			client.Search.WithQuery("unknown"),
		)
		require.NoError(t, err)
		require.False(t, res.IsError(), res.String())

		require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
		require.NoError(t, res.Body.Close())
		require.Equal(t, 0, out.Hits.Total.Value)
	}
}

func TestPreset_withDefaults(t *testing.T) {
	if israce.Enabled {
		t.Skip("elastic tests can't run with race detector due to https://github.com/elastic/go-elasticsearch/issues/147")
	}

	p := elastic.Preset()
	c, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	cfg := elasticsearch.Config{
		Addresses:    []string{fmt.Sprintf("http://%s", c.DefaultAddress())},
		DisableRetry: true,
	}

	_, err = elasticsearch.NewClient(cfg)
	require.NoError(t, err)
}
