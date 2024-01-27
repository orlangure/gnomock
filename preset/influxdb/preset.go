// Package influxdb includes InfluxDB implementation of Gnomock Preset
// interface. This Preset can be passed to gnomock.Start() function to create a
// configured InfluxDB container to use in tests.
//
// Currently this preset only supports InfluxDB 2.x. If you are interested in
// earlier versions, please consider opening an issue.
package influxdb

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	defaultVersion   = "2.7.6-alpine"
	defaultPort      = 8086
	defaultUsername  = "gnomock"
	defaultPassword  = "gnomock-password"
	defaultOrg       = "gnomock-org"
	defaultBucket    = "gnomock-bucket"
	defaultAuthToken = "gnomock-influxdb-token" //nolint:gosec
)

func init() {
	registry.Register("influxdb", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock InfluxDB preset. This preset includes a
// InfluxDB specific healthcheck function and default InfluxDB image and port.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for InfluxDB.
type P struct {
	Version   string `json:"version"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Org       string `json:"org"`
	Bucket    string `json:"bucket"`
	AuthToken string `json:"auth_token"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/influxdb:%s", p.Version)
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
		gnomock.WithEnv("DOCKER_INFLUXDB_INIT_MODE=setup"),
		gnomock.WithEnv("DOCKER_INFLUXDB_INIT_USERNAME=" + p.Username),
		gnomock.WithEnv("DOCKER_INFLUXDB_INIT_PASSWORD=" + p.Password),
		gnomock.WithEnv("DOCKER_INFLUXDB_INIT_ORG=" + p.Org),
		gnomock.WithEnv("DOCKER_INFLUXDB_INIT_BUCKET=" + p.Bucket),
		gnomock.WithEnv("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=" + p.AuthToken),
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}

	if p.Username == "" {
		p.Username = defaultUsername
	}

	if p.Password == "" {
		p.Password = defaultPassword
	}

	if p.Org == "" {
		p.Org = defaultOrg
	}

	if p.Bucket == "" {
		p.Bucket = defaultBucket
	}

	if p.AuthToken == "" {
		p.AuthToken = defaultAuthToken
	}
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	addr := fmt.Sprintf("http://%s", c.DefaultAddress())
	client := influxdb2.NewClient(addr, p.AuthToken)

	defer client.Close()

	h, err := client.Health(ctx)
	if err != nil {
		return fmt.Errorf("influxdb Health() return error: %w", err)
	}

	if h.Status != domain.HealthCheckStatusPass {
		return fmt.Errorf("unexpected health status '%s'", h.Status)
	}

	buckets, err := client.BucketsAPI().FindBucketsByOrgName(ctx, p.Org)
	if err != nil {
		return fmt.Errorf("can't list influxdb buckets: %w", err)
	}

	if len(*buckets) == 0 {
		return fmt.Errorf("no buckets exist")
	}

	return nil
}
