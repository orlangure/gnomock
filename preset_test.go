package gnomock_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/health"
	"github.com/orlangure/gnomock/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestPreset_parallel(t *testing.T) {
	t.Parallel()

	containers, err := gnomock.InParallel().
		Start(&testutil.TestPreset{Img: testutil.TestImage}).
		Start(&testutil.TestPreset{Img: testutil.TestImage}).
		Start(&testutil.TestPreset{Img: testutil.TestImage}).
		Start(&testutil.TestPreset{Img: testutil.TestImage}).
		Start(&testutil.TestPreset{Img: testutil.TestImage}).
		Go()

	defer func() { require.NoError(t, gnomock.Stop(containers...)) }()

	require.NoError(t, err)
	require.Len(t, containers, 5)

	ctx := context.Background()

	for _, c := range containers {
		require.NoError(t, health.HTTPGet(ctx, c.Address("web80")))
		require.NoError(t, health.HTTPGet(ctx, c.Address("web8080")))
	}
}

func TestPreset(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	container, err := gnomock.Start(p)

	defer func(c *gnomock.Container) {
		require.NoError(t, gnomock.Stop(c))
	}(container)

	require.NoError(t, err)
}

func TestPreset_containerRemainsIfDebug(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	container, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Second*30),
		gnomock.WithDebugMode(),
		gnomock.WithHealthCheck(failingHealthcheck),
	)

	require.Error(t, err)
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

	p := &testutil.TestPreset{Img: testutil.TestImage}
	originalContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock"),
		gnomock.WithDebugMode(),
	)
	require.NoError(t, err)

	newContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock"),
		gnomock.WithDebugMode(),
	)
	require.NoError(t, err)

	require.Error(t, gnomock.Stop(originalContainer))
	require.NoError(t, gnomock.Stop(newContainer))
}

func TestPreset_reusableContainerSucceeds(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	originalContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock-reuse-success"),
		gnomock.WithDebugMode(),
	)
	require.NoError(t, err)

	newContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock-reuse-success"),
		gnomock.WithDebugMode(),
		gnomock.WithContainerReuse(),
	)
	require.NoError(t, err)

	require.Equal(t, originalContainer.ID, newContainer.ID)
	require.NoError(t, gnomock.Stop(newContainer))
}

func TestPreset_reusableContainerFailsWithoutName(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	_, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerReuse(),
		gnomock.WithDebugMode(),
	)
	require.EqualError(t, err, "can't start container: container name is required when container reuse is enabled")
}

func TestPreset_reusableContainerResets(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	originalContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock-reuse-reset-no-error"),
	)
	require.NoError(t, err)

	resetCalled := false
	c, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock-reuse-reset-no-error"),
		gnomock.WithContainerReuse(),
		gnomock.WithContainerReset(func(c *gnomock.Container) error {
			resetCalled = true
			return nil
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, resetCalled)
	require.Equal(t, originalContainer.DefaultAddress(), c.DefaultAddress())
}

func TestPreset_reusableContainerResetsWithError(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	originalContainer, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock-reuse-reset-with-error"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, gnomock.Stop(originalContainer)) })

	c, err := gnomock.Start(
		p,
		gnomock.WithTimeout(time.Minute),
		gnomock.WithContainerName("gnomock-reuse-reset-with-error"),
		gnomock.WithContainerReuse(),
		gnomock.WithContainerReset(func(c *gnomock.Container) error {
			return fmt.Errorf("unable to reset")
		}),
	)
	require.ErrorContains(t, err, "unable to reset")
	require.Nil(t, c)
}

func TestPreset_customNamedPorts(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	presetPorts := p.Ports()
	pr := presetPorts["web80"]
	pr.HostPort = 23080
	presetPorts["web80"] = pr

	container, err := gnomock.Start(
		p,
		gnomock.WithCustomNamedPorts(presetPorts),
		gnomock.WithDebugMode(),
	)

	t.Cleanup(func() { require.NoError(t, gnomock.Stop(container)) })
	require.NoError(t, err)
	require.Equal(t, 23080, container.Ports.Get("web80").Port)
}
