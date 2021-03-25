package gnomock

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	// this test cannot run in parallel with other tests since it modifies the
	// environment, which affects other tests
	t.Run("fails with ErrEnvClient", func(t *testing.T) {
		currentHost := os.Getenv("DOCKER_HOST")

		defer func() {
			_ = os.Setenv("DOCKER_HOST", currentHost)
		}()

		_ = os.Setenv("DOCKER_HOST", "example.com")

		gg, err := newG(false)
		require.NoError(t, err)

		client, err := gg.dockerConnect()
		require.True(t, errors.Is(err, ErrEnvClient))
		require.Nil(t, client)
	})
}

func TestWaitForContainerNetwork(t *testing.T) {
	t.Parallel()

	// this test starts a regular container using existing API, and then uses
	// it to test specific error flows
	namedPorts := NamedPorts{
		"web80":   TCP(80),
		"web8080": TCP(8080),
	}
	container, err := StartCustom(
		"docker.io/orlangure/gnomock-test-image", namedPorts,
		WithTimeout(time.Second*10),
	)
	id, _ := parseID(container.ID)

	require.NoError(t, err)
	require.NotNil(t, container)

	gg, err := newG(false)
	require.NoError(t, err)

	d, err := gg.dockerConnect()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("fails after context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		_, err = d.waitForContainerNetwork(ctx, id, namedPorts)
		require.EqualError(t, err, "container network is unavailable after timeout")
	})

	t.Run("fails with wrong container id", func(t *testing.T) {
		_, err = d.waitForContainerNetwork(ctx, "wrong-id", namedPorts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "No such container")
	})

	t.Run("returns ErrPortNotFound for wrong port number", func(t *testing.T) {
		_, err := d.waitForContainerNetwork(ctx, id, DefaultTCP(42))
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrPortNotFound), err.Error())
	})
}
