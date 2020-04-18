// Package localstack provides a Gnomock Preset for localstack project
// (https://github.com/localstack/localstack). It allows to easily setup local
// AWS stack for testing
package localstack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/orlangure/gnomock"
)

const (
	webPort = "web"

	// APIPort should be used to configure AWS SDK endpoint
	APIPort = "api"
)

// Preset creates a new localstack preset to use with gnomock.Start. See
// package docs for a list of exposed ports and services. It is legal to not
// provide any services using WithServices options, but in such case a new
// localstack container will be useless.
//
// This Preset cannot be used with localstack image prior to 0.11.0
func Preset(opts ...Option) gnomock.Preset {
	p := &preset{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type preset struct {
	services []Service

	s3Path string
}

// Image returns an image that should be pulled to create this container
func (p *preset) Image() string {
	return "docker.io/localstack/localstack"
}

// Ports returns ports that should be used to access this container
func (p *preset) Ports() gnomock.NamedPorts {
	return gnomock.NamedPorts{
		webPort: {Protocol: "tcp", Port: 8080},
		APIPort: {Protocol: "tcp", Port: 4566},
	}
}

// Options returns a list of options to configure this container
func (p *preset) Options() []gnomock.Option {
	svcStrings := make([]string, len(p.services))
	for i, svc := range p.services {
		svcStrings[i] = string(svc)
	}

	svcEnv := strings.Join(svcStrings, ",")

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck(svcStrings)),
		gnomock.WithStartTimeout(time.Second * 60 * 2),
		gnomock.WithWaitTimeout(time.Second * 60),
		gnomock.WithEnv("SERVICES=" + svcEnv),
		gnomock.WithInit(p.initf()),
	}

	return opts
}

func (p *preset) healthcheck(services []string) gnomock.HealthcheckFunc {
	return func(c *gnomock.Container) (err error) {
		addr := fmt.Sprintf("http://%s/health", c.Address(webPort))

		res, err := http.Get(addr) //nolint:gosec
		if err != nil {
			return err
		}

		defer func() {
			closeErr := res.Body.Close()
			if err == nil && closeErr != nil {
				err = closeErr
			}
		}()

		var hr healthResponse

		decoder := json.NewDecoder(res.Body)

		err = decoder.Decode(&hr)
		if err != nil {
			return err
		}

		if len(hr.Services) < len(services) {
			return fmt.Errorf(
				"not enough active services: want %d got %d [%s]",
				len(services), len(hr.Services), hr.Services,
			)
		}

		for _, service := range services {
			status := hr.Services[service]
			if status != "running" {
				return fmt.Errorf("service '%s' is not running", service)
			}
		}

		return nil
	}
}

type healthResponse struct {
	Services map[string]string `json:"services"`
}

func (p *preset) initf() gnomock.InitFunc {
	return func(c *gnomock.Container) error {
		for _, s := range p.services {
			if s == S3 {
				err := p.initS3(c)
				if err != nil {
					return fmt.Errorf("can't init s3 storage: %w", err)
				}
			}
		}

		return nil
	}
}
