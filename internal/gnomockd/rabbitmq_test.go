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

// nolint:funlen,bodyclose
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

	q, err := ch.QueueDeclare(
		"gnomock",
		false, // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	require.NoError(t, err)

	msgBody := []byte("hello from Gnomock!")
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msgBody,
		},
	)
	require.NoError(t, err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	require.NoError(t, err)

	m := <-msgs
	require.Equal(t, msgBody, m.Body)

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
