// Package rabbitmq provides a Gnomock Preset for RabbitMQ.
package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/orlangure/gnomock"
	"github.com/streadway/amqp"
)

// ManagementPort is a name of the port exposed by RabbitMQ management plugin.
// This port is only available when an appropriate version of RabbitMQ docker
// image is used. See `Preset` docs for more info.
const ManagementPort = "management"

const defaultUser = "guest"
const defaultPassword = "guest"
const defaultVersion = "alpine"
const defaultPort = 5672
const managementPort = 15672

// Message is a single message sent to RabbitMQ.
type Message struct {
	Queue       string `json:"queue"`
	ContentType string `json:"contentType"`
	StringBody  string `json:"stringBody"`
	Body        []byte `json:"body"`
}

// Preset creates a new Gmomock RabbitMQ preset. This preset includes a
// RabbitMQ specific healthcheck function and default RabbitMQ image and port.
//
// By default, this preset does not use RabbitMQ Management plugin. To enable
// it, use one of the management tags with `WithVersion` option. Management
// port will be accessible using `container.Port(rabbitmq.ManagementPort)`. See
// https://hub.docker.com/_/rabbitmq/?tab=tags for a list of available tags.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of RabbitMQ.
type P struct {
	User          string    `json:"user"`
	Password      string    `json:"password"`
	Version       string    `json:"version"`
	Messages      []Message `json:"messages"`
	MessagesFiles []string  `json:"messages_files"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/rabbitmq:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	namedPorts := gnomock.DefaultTCP(defaultPort)

	if p.isManagement() {
		namedPorts[ManagementPort] = gnomock.Port{Protocol: "tcp", Port: managementPort}
	}

	return namedPorts
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
	}

	if p.User != "" && p.Password != "" {
		opts = append(
			opts,
			gnomock.WithEnv("RABBITMQ_DEFAULT_USER="+p.User),
			gnomock.WithEnv("RABBITMQ_DEFAULT_PASS="+p.Password),
		)
	}

	if len(p.Messages)+len(p.MessagesFiles) > 0 {
		opts = append(opts, gnomock.WithInit(p.initf))
	}

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	conn, err := p.connect(c)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("can't close connection: %w", err)
	}

	if p.isManagement() {
		addr := c.Address(ManagementPort)
		url := fmt.Sprintf("http://%s/api/overview", addr)

		// any non-err response is valid, it is most likely 401 Unauthorized
		resp, err := http.Get(url) // nolint:gosec
		if err != nil {
			return err
		}

		_ = resp.Body.Close()
	}

	return nil
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}

	if p.User == "" && p.Password == "" {
		p.User = defaultUser
		p.Password = defaultPassword
	}
}

func (p *P) loadFiles() error {
	if len(p.MessagesFiles) > 0 {
		for _, fName := range p.MessagesFiles {
			msgs, err := p.loadMessagesFromFile(fName)
			if err != nil {
				return fmt.Errorf("can't read messages from file '%s': %w", fName, err)
			}

			p.Messages = append(p.Messages, msgs...)
		}
	}

	return nil
}

func declareQueues(ch *amqp.Channel, qs []string) error {
	for _, queue := range qs {
		_, err := ch.QueueDeclare(queue, false, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("can't open queue '%s': %w", queue, err)
		}
	}

	return nil
}

func (p *P) initf(ctx context.Context, c *gnomock.Container) (err error) {
	conn, err := p.connect(c)
	if err != nil {
		return fmt.Errorf("can't connect to rabbitmq: %w", err)
	}

	defer func() {
		closeErr := conn.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if err := p.loadFiles(); err != nil {
		return err
	}

	messagesByQueue := make(map[string][]Message)
	for _, m := range p.Messages {
		messagesByQueue[m.Queue] = append(messagesByQueue[m.Queue], m)
	}

	queues := make([]string, len(messagesByQueue))
	for q := range messagesByQueue {
		queues = append(queues, q)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("can't open channel: %w", err)
	}

	defer func() {
		closeErr := ch.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if err := declareQueues(ch, queues); err != nil {
		return err
	}

	for queue, messages := range messagesByQueue {
		if err := p.sendMessagesIntoQueue(ch, queue, messages); err != nil {
			return fmt.Errorf("can't send messages into queue '%s': %w", queue, err)
		}
	}

	return nil
}

func (p *P) isManagement() bool {
	return strings.Contains(p.Version, "management")
}

// nolint:gosec
func (p *P) loadMessagesFromFile(fName string) (msgs []Message, err error) {
	f, err := os.Open(fName)
	if err != nil {
		return nil, fmt.Errorf("can't open messages file '%s': %w", fName, err)
	}

	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	decoder := json.NewDecoder(f)

	for {
		var m Message

		err = decoder.Decode(&m)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("can't read message from file '%s': %w", fName, err)
		}

		msgs = append(msgs, m)
	}

	return msgs, nil
}

func (p *P) connect(c *gnomock.Container) (*amqp.Connection, error) {
	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d", p.User, p.Password, c.Host, c.DefaultPort()))
}

func (p *P) sendMessagesIntoQueue(ch *amqp.Channel, q string, msgs []Message) (err error) {
	for _, m := range msgs {
		var body []byte
		if m.Body != nil {
			body = m.Body
		} else {
			body = []byte(m.StringBody)
		}

		if err := ch.Publish(
			"",
			q,
			false,
			false,
			amqp.Publishing{
				ContentType: m.ContentType,
				Body:        body,
			},
		); err != nil {
			return fmt.Errorf("publish message failed: %w", err)
		}
	}

	return nil
}
