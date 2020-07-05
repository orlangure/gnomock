# Gnomock RabbitMQ

Gnomock RabbitMQ is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real RabbitMQ message queue, without mocks.

```go
package rabbitmq_test

import (
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/rabbitmq"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

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
```
