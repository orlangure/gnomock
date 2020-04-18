// Package redis includes Redis implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.StartPreset function to create a
// configured Redis container to use in tests
package redis

import (
	"fmt"

	redisclient "github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
)

// Preset creates a new Gmomock Redis preset. This preset includes a Redis
// specific healthcheck function, default Redis image and port, and allows to
// optionally set up initial state
func Preset(opts ...Option) gnomock.Preset {
	p := &preset{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// preset is a Gnomock Preset implementation for preset storage
type preset struct {
	initialValues map[string]interface{}
}

// Image returns an image that should be pulled to create this container
func (r *preset) Image() string {
	return "docker.io/library/redis"
}

// Ports returns ports that should be used to access this container
func (r *preset) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(6379)
}

// Options returns a list of options to configure this container
func (r *preset) Options() []gnomock.Option {
	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
	}

	if r.initialValues != nil {
		initf := func(c *gnomock.Container) error {
			addr := c.Address(gnomock.DefaultPort)
			client := redisclient.NewClient(&redisclient.Options{Addr: addr})

			for k, v := range r.initialValues {
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

func healthcheck(c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)
	client := redisclient.NewClient(&redisclient.Options{Addr: addr})
	_, err := client.Ping().Result()

	return err
}
