// Package redis includes Redis implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.Start() function to create a configured
// Redis container to use in tests.
package redis

import (
	"context"
	"fmt"

	redisclient "github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const defaultVersion = "6.0.9"

func init() {
	registry.Register("redis", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Redis preset. This preset includes a Redis
// specific healthcheck function, default Redis image and port, and allows to
// optionally set up initial state.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for Redis storage.
type P struct {
	Values  map[string]interface{} `json:"values"`
	Version string                 `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/redis:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(6379)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
	}

	if p.Values != nil {
		initf := func(ctx context.Context, c *gnomock.Container) error {
			addr := c.Address(gnomock.DefaultPort)
			client := redisclient.NewClient(&redisclient.Options{Addr: addr})

			for k, v := range p.Values {
				err := client.Set(k, v, 0).Err()
				if err != nil {
					return fmt.Errorf("can't set '%s'='%v': %w", k, v, err)
				}
			}

			return nil
		}

		opts = append(opts, gnomock.WithInit(initf))
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func healthcheck(_ context.Context, c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)
	client := redisclient.NewClient(&redisclient.Options{Addr: addr})
	_, err := client.Ping().Result()

	return err
}
