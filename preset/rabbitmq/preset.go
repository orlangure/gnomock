// Package rabbitmq provides a Gnomock Preset for RabbitMQ
package rabbitmq

import (
	"context"
	"fmt"

	"github.com/orlangure/gnomock"
	"github.com/streadway/amqp"
)

const defaultUser = "guest"
const defaultPassword = "guest"
const defaultVersion = "alpine"
const defaultPort = 5672

// Preset creates a new Gmomock RabbitMQ preset. This preset includes a RabbitMQ
// specific healthcheck function and default RabbitMQ image and port.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of RabbitMQ.
type P struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Version  string `json:"version"`
}

// Image returns an image that should be pulled to create this container
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/rabbitmq:%s", p.Version)
}

// Ports returns ports that should be used to access this container
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container
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

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%d", p.User, p.Password, c.Host, c.DefaultPort())

	conn, err := amqp.Dial(uri)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("can't close connection: %w", err)
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
