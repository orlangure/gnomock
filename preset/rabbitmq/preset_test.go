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

func TestPreset(t *testing.T) {
	t.Parallel()

	messages := []rabbitmq.Message{
		{
			Queue:       "events",
			ContentType: "text/plain",
			StringBody:  "order: 1",
		},
		{
			Queue:       "alerts",
			ContentType: "text/plain",
			StringBody:  "CPU: 92",
		},
	}

	byteMessages := []rabbitmq.Message{
		{
			Queue:       "events",
			ContentType: "text/binary", // non-existent format for test
			Body:        []byte{54, 23, 12, 76, 54},
		},
		{
			Queue:       "alerts",
			ContentType: "text/binary", // non-existent format for test
			Body:        []byte{75, 12, 8, 42, 12},
		},
	}

	// gnomock setup
	p := rabbitmq.Preset(
		rabbitmq.WithUser("gnomock", "strong-password"),
		rabbitmq.WithMessages(messages...),
		rabbitmq.WithMessages(byteMessages...),
		rabbitmq.WithMessagesFile("./testdata/messages.json"),
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

	// ===================================
	// Test for string and binary messages
	// ===================================
	msgs, err = ch.Consume("events", "", true, false, false, false, nil)
	require.NoError(t, err)

	m, ok := <-msgs
	require.Equal(t, true, ok)
	require.Equal(t, []byte(messages[0].StringBody), m.Body)

	m, ok = <-msgs
	require.Equal(t, true, ok)
	require.Equal(t, byteMessages[0].Body, m.Body)

	msgs, err = ch.Consume("alerts", "", true, false, false, false, nil)
	require.NoError(t, err)

	m, ok = <-msgs
	require.Equal(t, true, ok)
	require.Equal(t, []byte(messages[1].StringBody), m.Body)

	m, ok = <-msgs
	require.Equal(t, true, ok)
	require.Equal(t, byteMessages[1].Body, m.Body)
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

	resp, err := http.Get(url)
	require.NoError(t, err)

	defer require.NoError(t, resp.Body.Close())

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := rabbitmq.Preset()
	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	uri := fmt.Sprintf(
		"amqp://%s:%s@%s",
		"guest", "guest",
		container.DefaultAddress(),
	)
	conn, err := amqp.Dial(uri)
	require.NoError(t, err)

	defer func() { require.NoError(t, conn.Close()) }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	require.NoError(t, ch.Close())
}
