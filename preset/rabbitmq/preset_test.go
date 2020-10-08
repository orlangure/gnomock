// +build !nopreset

package rabbitmq_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/rabbitmq"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

// nolint:funlen
func TestPreset(t *testing.T) {
	t.Parallel()

	// gnomock setup
	p := rabbitmq.Preset(
		rabbitmq.WithUser("gnomock", "strong-password"),
	)

	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	// actual test code
	uri := fmt.Sprintf(
		"amqp://%s:%s@%s",
		"gnomock", "strong-password",
		container.DefaultAddress(),
	)
	conn, err := amqp.Dial(uri)
	require.NoError(t, err)

	defer func() { require.NoError(t, conn.Close()) }()

	ch, err := conn.Channel()
	require.NoError(t, err)

	defer func() { require.NoError(t, ch.Close()) }()

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
}

func TestPreset_withManagement(t *testing.T) {
	t.Parallel()

	// gnomock setup
	p := rabbitmq.Preset(
		rabbitmq.WithUser("gnomock", "strong-password"),
		rabbitmq.WithVersion("management-alpine"),
	)

	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	addr := container.Address(rabbitmq.ManagementPort)
	url := fmt.Sprintf("http://%s/api/overview", addr)

	resp, err := http.Get(url) // nolint:gosec
	require.NoError(t, err)

	defer require.NoError(t, resp.Body.Close())

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPreset_withMessages(t *testing.T) {
	t.Parallel()

	messages := []rabbitmq.Message{
		{
			Queue:       "events",
			ContentType: "text/plain",
			Body:        "order: 1",
		},
		{
			Queue:       "alerts",
			ContentType: "text/plain",
			Body:        "CPU: 92",
		},
	}

	// gnomock setup
	p := rabbitmq.Preset(
		rabbitmq.WithUser("gnomock", "strong-password"),
		rabbitmq.WithQueues("topic-1", "topic-2"),
		rabbitmq.WithMessages(messages...),
	)

	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	// actual test code
	uri := fmt.Sprintf(
		"amqp://%s:%s@%s",
		"gnomock", "strong-password",
		container.DefaultAddress(),
	)
	conn, err := amqp.Dial(uri)
	require.NoError(t, err)

	defer func() { require.NoError(t, conn.Close()) }()

	ch, err := conn.Channel()
	require.NoError(t, err)

	defer func() { require.NoError(t, ch.Close()) }()

	msgs, err := ch.Consume("events", "", true, false, false, false, nil)
	require.NoError(t, err)

	m, ok := <-msgs
	require.Equal(t, true, ok)
	require.Equal(t, []byte(messages[0].Body), m.Body)

	msgs, err = ch.Consume("alerts", "", true, false, false, false, nil)
	require.NoError(t, err)

	m, ok = <-msgs
	require.Equal(t, true, ok)
	require.Equal(t, []byte(messages[1].Body), m.Body)
}
