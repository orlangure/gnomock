// Package testutil includes utilities used in test code of other packages.
package testutil

import (
	"context"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/health"
)

// Test preset uses the following values:
const (
	TestImage    = "docker.io/orlangure/gnomock-test-image"
	GoodPort80   = 80
	GoodPort8080 = 8080
	BadPort      = 8000
)

// TestPreset is a preset used for tests.
type TestPreset struct {
	Img string
}

// Image returns a canonical docker image used to setup this Preset.
func (t *TestPreset) Image() string {
	return t.Img
}

// Ports returns a group of ports exposed by this Preset, where every port
// is given a unique name.
//
// This Preset returns port 80 by "default" name.
func (t *TestPreset) Ports() gnomock.NamedPorts {
	namedPorts := gnomock.NamedPorts{
		"web80":   gnomock.TCP(GoodPort80),
		"web8080": gnomock.TCP(GoodPort8080),
	}

	return namedPorts
}

// Options returns a list of Option functions that allow to setup this
// Preset implementation. This test preset always returns a failing healthcheck.
func (t *TestPreset) Options() []gnomock.Option {
	return []gnomock.Option{
		gnomock.WithHealthCheck(Healthcheck),
	}
}

// Healthcheck returns an error if either of ports 80 and 8080 of test preset
// returns an unexpected response.
func Healthcheck(ctx context.Context, c *gnomock.Container) error {
	if err := health.HTTPGet(ctx, c.Address("web80")); err != nil {
		return err
	}

	if err := health.HTTPGet(ctx, c.Address("web8080")); err != nil {
		return err
	}

	return nil
}
