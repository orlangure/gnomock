package gnomockd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/rabbitmq"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

func TestRabbitMQ(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := ioutil.ReadFile("./testdata/rabbitmq.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/rabbitmq", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	uri := fmt.Sprintf(
		"amqp://%s:%s@%s",
		"guest", "guest",
		c.DefaultAddress(),
	)
	conn, err := amqp.Dial(uri)
	require.NoError(t, err)

	ch, err := conn.Channel()
	require.NoError(t, err)

	msgs, err := ch.Consume(
		"events", // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	require.NoError(t, err)

	m := <-msgs
	require.Equal(t, []byte("something"), m.Body)

	m = <-msgs
	require.Equal(t, "foobar", string(m.Body))

	m = <-msgs
	require.Equal(t, []byte("something else"), m.Body)

	msgs, err = ch.Consume(
		"alerts", // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	require.NoError(t, err)

	m = <-msgs
	require.Equal(t, []byte("high cpu"), m.Body)

	m = <-msgs
	require.Equal(t, []byte("memory pressure"), m.Body)

	m = <-msgs
	require.Equal(t, "gnomock", string(m.Body))

	require.NoError(t, ch.Close())
	require.NoError(t, conn.Close())

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
