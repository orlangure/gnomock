// Package splunk includes Splunk Enterprise implementation of Gnomock Preset
// interface. This Preset can be passed to gnomock.StartPreset function to
// create a configured Splunk container to use in tests.
//
// Splunk image is relatively heavy (larger than 1.5GB), and its startup time
// is longer than usual. Using this container may make the tests much longer.
package splunk

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	// CollectorPort is the name of a port exposed by Splunk Collector
	CollectorPort string = "collector"

	// APIPort is the name of a port exposed by Splunk API
	APIPort string = "api"

	// WebPort is the name of a port exposed by Splunk web UI
	WebPort string = "web"
)

const defaultVersion = "latest"

func init() {
	registry.Register("splunk", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gnomock Splunk preset. This preset includes a Splunk
// specific healthcheck function, default Splunk image and ports, and allows to
// optionally ingest initial logs
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of Splunk
type P struct {
	Values        []Event `json:"values"`
	ValuesFile    string  `json:"values_file"`
	AcceptLicense bool    `json:"accept_license"`
	AdminPassword string  `json:"admin_password"`
	Version       string  `json:"version"`
}

// Image returns an image that should be pulled to create this container
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/splunk/splunk:%s", p.Version)
}

// Ports returns ports that should be used to access this container
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.NamedPorts{
		CollectorPort: gnomock.TCP(8088),
		APIPort:       gnomock.TCP(8089),
		WebPort:       gnomock.TCP(8000),
	}
}

// Options returns a list of options to configure this container
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck(p.AdminPassword)),
		gnomock.WithEnv("SPLUNK_PASSWORD=" + p.AdminPassword),
	}

	if p.AcceptLicense {
		opts = append(
			opts,
			gnomock.WithEnv("SPLUNK_START_ARGS=--accept-license"),
		)
	}

	if p.Values != nil || p.ValuesFile != "" {
		init := p.initf()
		opts = append(opts, gnomock.WithInit(init))
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func healthcheck(password string) gnomock.HealthcheckFunc {
	return func(ctx context.Context, c *gnomock.Container) (err error) {
		err = checkAPI(ctx, c, password)
		if err != nil {
			return err
		}

		err = checkHEC(ctx, c)
		if err != nil {
			return err
		}

		return nil
	}
}

func checkAPI(ctx context.Context, c *gnomock.Container, password string) error {
	post := requestWithAuth(ctx, http.MethodPost, password, false)
	uri := fmt.Sprintf("https://%s/services/auth/login", c.Address(APIPort))

	data := url.Values{}
	data.Add("username", "admin")
	data.Add("password", password)
	data.Add("output_mode", "json")
	buf := bytes.NewBufferString(data.Encode())

	_, err := post(uri, buf)

	return err
}

func checkHEC(ctx context.Context, c *gnomock.Container) error {
	get := requestWithAuth(ctx, http.MethodGet, "", false)
	uri := fmt.Sprintf("https://%s/services/collector/health", c.Address(CollectorPort))

	_, err := get(uri, bytes.NewBufferString(""))

	return err
}

func insecureClient() http.Client {
	return http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
}
