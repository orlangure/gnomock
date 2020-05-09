// +build !nognomock

package gnomock_test

import (
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/stretchr/testify/require"
)

func TestPreset_parallel(t *testing.T) {
	t.Parallel()

	containers, err := gnomock.InParallel().
		Start(&testPreset{testImage}, gnomock.WithHealthCheck(healthcheck)).
		Start(&testPreset{testImage}, gnomock.WithHealthCheck(healthcheck)).
		Start(&testPreset{testImage}, gnomock.WithHealthCheck(healthcheck)).
		Start(&testPreset{testImage}, gnomock.WithHealthCheck(healthcheck)).
		Start(&testPreset{testImage}, gnomock.WithHealthCheck(healthcheck)).
		Go()

	defer func() { require.NoError(t, gnomock.Stop(containers...)) }()

	require.NoError(t, err)
	require.Len(t, containers, 5)

	for _, c := range containers {
		require.NoError(t, callRoot("http://"+c.Address("web80")))
		require.NoError(t, callRoot("http://"+c.Address("web8080")))
	}
}

func TestPreset(t *testing.T) {
	t.Parallel()

	p := &testPreset{testImage}
	container, err := gnomock.Start(p)

	defer func(c *gnomock.Container) {
		require.NoError(t, gnomock.Stop(c))
	}(container)

	// by default, testPreset always fails its healthcheck
	require.Error(t, err)

	container, err = gnomock.Start(p, gnomock.WithHealthCheck(healthcheck))

	defer func(c *gnomock.Container) {
		require.NoError(t, gnomock.Stop(c))
	}(container)

	// if we override healthcheck, container start correctly
	require.NoError(t, err)
}

func TestPreset_overrideTag(t *testing.T) {
	t.Parallel()

	p := &testPreset{testImage + ":latest"}
	container, err := gnomock.Start(p, gnomock.WithTag("bad"))

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.Error(t, err)
}

type testPreset struct {
	image string
}

// Image returns a canonical docker image used to setup this Preset
func (t *testPreset) Image() string {
	return t.image
}

// Ports returns a group of ports exposed by this Preset, where every port
// is given a unique name.
//
// This Preset returns port 80 by "default" name
func (t *testPreset) Ports() gnomock.NamedPorts {
	namedPorts := gnomock.NamedPorts{
		"web80":   gnomock.TCP(goodPort80),
		"web8080": gnomock.TCP(goodPort8080),
	}

	return namedPorts
}

// Options returns a list of Option functions that allow to setup this
// Preset implementation. This test preset always returns a failing healthcheck
func (t *testPreset) Options() []gnomock.Option {
	return []gnomock.Option{
		gnomock.WithHealthCheck(func(*gnomock.Container) error {
			return fmt.Errorf("this container should not start")
		}),
	}
}
