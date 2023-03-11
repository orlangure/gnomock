// Package azurite provides a Gnomock Preset for azurite project
// It allows to easily setup local
// Blobstorage for testing
package azurite

import (
	"context"
	"fmt"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	BlobServicePort  = "blob"
	QueueServicePort = "queue"
	TableServicePort = "table"
)

const (
	ConnectionStringFormat = "DefaultEndpointsProtocol=http;AccountName=%s;AccountKey=%s;BlobEndpoint=http://%s/%s;"
	AccountName            = "devstoreaccount1"
	AccountKey             = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	DefaultVersion         = "3.22.0"
)

func init() {
	registry.Register("azurite", func() gnomock.Preset { return &P{} })
}

// Preset creates a new azurite preset to use with gnomock.Start. See
// package docs for a list of exposed ports.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset blobstorage implementation.
type P struct {
	BlobstorePath string `json:"blobstore_path"`
	Version       string `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("mcr.microsoft.com/azure-storage/azurite:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.NamedPorts{
		BlobServicePort:  {Protocol: "tcp", Port: 10000},
		QueueServicePort: {Protocol: "tcp", Port: 10001},
		TableServicePort: {Protocol: "tcp", Port: 10002},
	}
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck()),
		gnomock.WithEnv(fmt.Sprintf("AZURITE_ACCOUNTS=%s:%s", AccountName, AccountKey)),
		gnomock.WithInit(p.initf()),
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = DefaultVersion
	}
}

func (p *P) healthcheck() gnomock.HealthcheckFunc {
	return func(ctx context.Context, c *gnomock.Container) (err error) {
		//needs implementation. unfortunately azurite does not offer a health endpoint
		//a netcat p.healtCheckAddress(c) -z could help alternatively
		_ = p.healthCheckAddress(c)
		return nil
	}
}

// healthCheckAddress returns the address of azurite blobstorage endpoint of a running
// azurite container.
func (p *P) healthCheckAddress(c *gnomock.Container) string {
	defaultPath := c.Address(BlobServicePort)

	return defaultPath
}

func (p *P) initf() gnomock.InitFunc {
	return func(ctx context.Context, c *gnomock.Container) error {
		err := p.initBlobstorage(c)
		if err != nil {
			return fmt.Errorf("can't init blob storage: %w", err)
		}

		return nil
	}
}
