package gnomockd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	"github.com/orlangure/gnomock/preset/kafka"
	kafkaclient "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

// nolint:bodyclose,funlen
func TestKafka(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := ioutil.ReadFile("./testdata/kafka.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/kafka", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	alertsReader := kafkaclient.NewReader(kafkaclient.ReaderConfig{
		Brokers: []string{c.Address(kafka.BrokerPort)},
		Topic:   "alerts",
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	m, err := alertsReader.ReadMessage(ctx)
	require.NoError(t, err)
	require.Equal(t, "CPU", string(m.Key))
	require.Equal(t, "92", string(m.Value))
	require.NoError(t, alertsReader.Close())

	eventsReader := kafkaclient.NewReader(kafkaclient.ReaderConfig{
		Brokers: []string{c.Address(kafka.BrokerPort)},
		Topic:   "events",
	})

	m, err = eventsReader.ReadMessage(ctx)
	require.NoError(t, err)
	require.Equal(t, "registration", string(m.Key))
	require.Equal(t, "abc", string(m.Value))

	m, err = eventsReader.ReadMessage(ctx)
	require.NoError(t, err)
	require.Equal(t, "order", string(m.Key))
	require.Equal(t, "42", string(m.Value))

	require.NoError(t, eventsReader.Close())

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
