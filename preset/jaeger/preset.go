// Package jaeger includes Jaeger implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.Start() function to create a configured
// Jaeger container to use in tests.
package jaeger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	// DefaultPort is the default port for Jaeger Query Service (UI)
	DefaultPort = 16686
	// CollectorHTTPPort is the default port for Jaeger Collector HTTP
	CollectorHTTPPort = 14268
	// HealthCheckPort is the default port for Jaeger Health Check
	HealthCheckPort = 14269

	defaultVersion = "1.57"
)

func init() {
	registry.Register("jaeger", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Jaeger preset. This preset includes a Jaeger
// specific healthcheck function, default Jaeger image and ports.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for Jaeger.
type P struct {
	Version string `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/jaegertracing/all-in-one:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.NamedPorts{
		"web":       gnomock.TCP(DefaultPort),
		"collector": gnomock.TCP(CollectorHTTPPort),
		"health":    gnomock.TCP(HealthCheckPort),
	}
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func healthcheck(ctx context.Context, c *gnomock.Container) error {
	addr := c.Address("health")
	url := fmt.Sprintf("http://%s/", addr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
