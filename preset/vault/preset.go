// Package vault includes vault implementation of Gnomock Preset
// interface. This Preset can be passed to gnomock.Start() function to create a
// configured vault container to use in tests.
package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	defaultVersion   = "1.13.0"
	defaultPort      = 8200
	defaultAuthToken = "gnomock-vault-token" //nolint:gosec
)

func init() {
	registry.Register("vault", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Vault preset. This preset includes a
// vault specific healthcheck function and default vault image and port.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for vault.
type P struct {
	Version   string   `json:"version"`
	AuthToken string   `json:"auth_token"`
	Auth      []Auth   `json:"auth"`
	Policies  []Policy `json:"policies"`
}

// Auth represents a vault authorization.
type Auth struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// Policy is a vault policy.
type Policy struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("hashicorp/vault:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithEnv("VAULT_DEV_ROOT_TOKEN_ID=" + p.AuthToken),
		gnomock.WithInit(p.initf()),
		gnomock.WithHealthCheck(p.healthcheck),
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}

	if p.AuthToken == "" {
		p.AuthToken = defaultAuthToken
	}
}

func (p *P) initf() gnomock.InitFunc {
	return func(ctx context.Context, c *gnomock.Container) error {
		cli, err := Client(c, p.AuthToken)
		if err != nil {
			return err
		}

		for _, m := range p.Auth {
			if err := cli.Sys().EnableAuthWithOptions(m.Path, &api.EnableAuthOptions{
				Type:        m.Type,
				Description: m.Path,
			}); err != nil {
				return fmt.Errorf("failed to enable %s auth on %s: %w", m.Type, m.Path, err)
			}
		}

		for _, pol := range p.Policies {
			if err := cli.Sys().PutPolicy(pol.Name, pol.Data); err != nil {
				return fmt.Errorf("failed to write policy %s: %w", pol.Name, err)
			}
		}

		return nil
	}
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	cli, err := Client(c, p.AuthToken)
	if err != nil {
		return err
	}

	_, err = cli.Sys().Health()

	return err
}

// Client creates a configured vault client for the provided container and token.
func Client(c *gnomock.Container, token string) (*api.Client, error) {
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = fmt.Sprintf("http://%s", c.DefaultAddress())

	vaultClient, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}

	vaultClient.SetToken(token)

	return vaultClient, nil
}

// CreateToken creates an additional access token with the provided policies. Use the same password you provided
// with the WithAuthToken option.
func CreateToken(c *gnomock.Container, rootToken string, policies ...string) (string, error) {
	cli, err := Client(c, rootToken)
	if err != nil {
		return "", err
	}

	tcr := api.TokenCreateRequest{
		Policies: policies,
	}

	s, err := cli.Auth().Token().Create(&tcr)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return s.Auth.ClientToken, nil
}
