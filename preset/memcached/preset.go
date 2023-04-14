// Package memcached includes Memcached implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.StartPreset function to create a
// configured Memcached container to use in tests
package memcached

import (
	"context"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const defaultVersion = "1.6.9"

func init() {
	registry.Register("memcached", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Memcached preset. This preset includes a Memcached
// specific healthcheck function, default Memcached image and port, and allows to
// optionally set up initial state.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for Memcached storage.
type P struct {
	Values     map[string]string `json:"values"`
	ByteValues map[string][]byte `json:"byteValues"`
	Version    string            `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/memcached:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(11211)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
	}

	if p.ByteValues != nil || p.Values != nil {
		initf := func(ctx context.Context, c *gnomock.Container) error {
			addr := c.Address(gnomock.DefaultPort)
			client := memcache.New(addr)

			if p.ByteValues != nil {
				for k, v := range p.ByteValues {
					err := client.Set(&memcache.Item{Key: k, Value: v, Expiration: 0})
					if err != nil {
						return fmt.Errorf("can't set '%s'='%v': %w", k, v, err)
					}
				}
			}

			if p.Values != nil {
				for k, v := range p.Values {
					err := client.Set(&memcache.Item{Key: k, Value: []byte(v), Expiration: 0})
					if err != nil {
						return fmt.Errorf("can't set '%s'='%v': %w", k, v, err)
					}
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
	client := memcache.New(addr)

	return client.Ping()
}
