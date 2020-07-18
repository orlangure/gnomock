// Package rabbitmq provides a Gnomock Preset for RabbitMQ.
package rabbitmq

import (
	"context"
	"fmt"
	"net/http"
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
	User     string `json:"user"`
	Password string `json:"password"`
	Version  string `json:"version"`
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

func (p *P) isManagement() bool {
	return strings.Contains(p.Version, "management")
}
