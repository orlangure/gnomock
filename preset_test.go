// +build !nognomock

package gnomock_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	container, err := gnomock.Start(p, gnomock.WithTimeout(time.Second*15))

	// by default, testPreset always fails its healthcheck
	require.Error(t, err)
	require.Contains(t, err.Error(), "this container should not start")

	defer func(c *gnomock.Container) {
		require.NoError(t, gnomock.Stop(c))
	}(container)

	container, err = gnomock.Start(p, gnomock.WithHealthCheck(healthcheck))

	defer func(c *gnomock.Container) {
		require.NoError(t, gnomock.Stop(c))
	}(container)

	// if we override healthcheck, container start correctly
	require.NoError(t, err)
}

func TestPreset_containerRemainsIfDebug(t *testing.T) {
	t.Parallel()

	p := &testPreset{testImage}
	container, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Second*15),
		gnomock.WithDebugMode(),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "this container should not start")
	require.NotNil(t, container)

	// if stopped without error, container existed at this point
	err = gnomock.Stop(container)
	require.NoError(t, err)

	// allow the container to actually stop
	time.Sleep(time.Second)

	// confirm it doesn't exist anymore
	err = gnomock.Stop(container)
	require.Error(t, err)
}

func TestPreset_duplicateContainerName(t *testing.T) {
	t.Parallel()

	p := &testPreset{testImage}
	originalContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Second*15),
		gnomock.WithContainerName("gnomock"),
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithDebugMode(),
	)
	require.NoError(t, err)

	newContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Second*15),
		gnomock.WithContainerName("gnomock"),
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithDebugMode(),
	)
	require.NoError(t, err)

	require.Error(t, gnomock.Stop(originalContainer))
	require.NoError(t, gnomock.Stop(newContainer))
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
		gnomock.WithHealthCheck(func(context.Context, *gnomock.Container) error {
			return fmt.Errorf("this container should not start")
		}),
	}
}
