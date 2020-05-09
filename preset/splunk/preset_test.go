// +build !nopreset

package splunk_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/splunk"
	"github.com/stretchr/testify/require"
)

func ExamplePreset() {
	events := []splunk.Event{
		{
			Event:      "action=foo",
			Index:      "events",
			Source:     "app",
			SourceType: "http",
			Time:       time.Now().UnixNano(),
		},
		{
			Event:      "action=bar",
			Index:      "events",
			Source:     "app",
			SourceType: "http",
			Time:       time.Now().UnixNano(),
		},
	}

	p := splunk.Preset(
		splunk.WithVersion("latest"),
		splunk.WithLicense(true),
		splunk.WithPassword("12345678"),
		splunk.WithValues(events),
		splunk.WithInitTimeout(time.Second*10),
	)

	// created container now includes two events in "events" index
	container, err := gnomock.Start(p)
	fmt.Println("error:", err)
	fmt.Println(len(container.Ports), "exposed ports:")

	defer func() {
		_ = gnomock.Stop(container)
	}()

	// Port numbers as well as container address are non-deterministic, so they
	// are skipped in this example. The usage would be:
	//
	//		container.Address("web")
	// 		container.Address("api")

	portNames := make([]string, 0)

	for portName := range container.Ports {
		portNames = append(portNames, portName)
	}

	sort.Strings(portNames)
	fmt.Println(portNames)

	// Output:
	// error: <nil>
	// 3 exposed ports:
	// [api collector web]
}

//nolint:funlen
func TestPreset(t *testing.T) {
	events := make([]splunk.Event, 1000)

	for i := range events {
		e := splunk.Event{
			Index:      "foo",
			Source:     "bar",
			SourceType: "baz",
			Event:      fmt.Sprintf("some=%d", i),
			Time:       time.Now().UnixNano(),
		}

		if i%3 == 0 {
			e.Event += fmt.Sprintf(" other=%d", i*2)
		}

		if i%5 == 0 {
			e.Event += fmt.Sprintf(" third=%d%d", i+1000, i*2000)
		}

		events[i] = e
	}

	p := splunk.Preset(
		splunk.WithVersion("8.0.2.1"),
		splunk.WithLicense(true),
		splunk.WithPassword("12345678"),
		splunk.WithValues(events),
		splunk.WithInitTimeout(time.Second*20),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()
	require.NoError(t, err)

	err = splunk.Ingest(context.Background(), c, "12345678", splunk.Event{
		Index:      "foo",
		Source:     "bar",
		SourceType: "baz",
		Event:      "separate event",
		Time:       time.Now().UnixNano(),
	})
	require.NoError(t, err)

	t.Run("initial values ingested", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}

		data := url.Values{}
		data.Add("search", "search index=foo some > 900 | stats count")
		data.Add("earliest", "-1d")
		data.Add("latest", "now")
		data.Add("output_mode", "json")
		buf := bytes.NewBufferString(data.Encode())

		addr := fmt.Sprintf("https://%s/services/search/jobs/export", c.Address(splunk.APIPort))
		req, err := http.NewRequest(http.MethodPost, addr, buf)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "12345678")
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		defer func() { require.NoError(t, res.Body.Close()) }()

		r := struct {
			Result struct {
				Count string `json:"count"`
			} `json:"result"`
		}{}

		bs, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(bs, &r))
		require.Equal(t, "99", r.Result.Count)
	})

	t.Run("context canceled", func(t *testing.T) {
		events := make([]splunk.Event, 1000)

		for i := range events {
			e := splunk.Event{
				Index:      "foo",
				Source:     "bar",
				SourceType: "baz",
				Event:      fmt.Sprintf("some=%d", i),
				Time:       time.Now().UnixNano(),
			}

			if i%3 == 0 {
				e.Event += fmt.Sprintf(" other=%d", i*2)
			}

			if i%5 == 0 {
				e.Event += fmt.Sprintf(" third=%d%d", i+1000, i*2000)
			}

			events[i] = e
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = splunk.Ingest(ctx, c, "12345678", events...)
		require.Truef(t, errors.Is(err, context.Canceled), "want context.Canceled, got %v", err)
	})
}
