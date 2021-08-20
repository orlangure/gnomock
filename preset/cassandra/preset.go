// Package cassandra includes Cassandra implementation of Gnomock Preset
// interface. This Preset can be passed to gnomock.Start() function to create a
// configured Cassandra container to use in tests.
//
// Cassandra containers always use cassandra/cassandra username/password pair,
// it is currently not possible to use different values.
package cassandra

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

// By default, Cassandra containers will use these values.
const (
	DefaultUser     = "cassandra"
	DefaultPassword = "cassandra"

	defaultVersion = "3"
	defaultPort    = 9042
)

func init() {
	registry.Register("cassandra", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Cassandra preset. This preset includes a
// Cassandra specific healthcheck function and default Cassandra image and
// port.
//
// Containers created using this preset should be accessed using
// cassandra/cassandra username/password pair.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for Cassandra.
type P struct {
	Version string `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/bitnami/cassandra:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	cluster := gocql.NewCluster(c.DefaultAddress())
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: DefaultUser,
		Password: DefaultPassword,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create a new session: %w", err)
	}

	session.Close()

	return nil
}
