package gnomock

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testImage = "docker.io/orlangure/gnomock-test-image"

func TestWaitForContainerNetwork(t *testing.T) {
	t.Parallel()

	// this test starts a regular container using existing API, and then uses
	// it to test specific error flows
	namedPorts := NamedPorts{
		"web80":   TCP(80),
		"web8080": TCP(8080),
	}
	container, err := StartCustom(
		testImage, namedPorts,
		WithTimeout(time.Minute),
	)
	require.NoError(t, err)
	require.NotNil(t, container)

	id, _ := parseID(container.ID)

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

func TestEnvAwareClone(t *testing.T) {
	// this test cannot run in parallel with other tests since it modifies the
	// environment, which affects other tests
	original := &Container{
		ID:      "foo",
		Host:    "bar",
		gateway: "gateway",
	}

	t.Run("same host returned in regular flow", func(t *testing.T) {
		cloned := envAwareClone(original)
		require.Equal(t, original.ID, cloned.ID)
		require.Equal(t, original.Host, cloned.Host)
	})

	t.Run("container gateway returned when no internal host available", func(t *testing.T) {
		currentEnv := os.Getenv("GNOMOCK_ENV")
		_ = os.Setenv("GNOMOCK_ENV", "gnomockd")

		defer func() {
			_ = os.Setenv("GNOMOCK_ENV", currentEnv)
		}()

		cloned := envAwareClone(original)
		require.Equal(t, original.ID, cloned.ID)
		require.Equal(t, original.gateway, cloned.Host)
	})
}

func TestCustomDockerHost(t *testing.T) {
	// this test cannot run in parallel with other tests since it modifies the
	// environment, which affects other tests
	t.Run("fails with misconfigured docker host", func(t *testing.T) {
		currentHost := os.Getenv("DOCKER_HOST")

		defer func() {
			_ = os.Setenv("DOCKER_HOST", currentHost)
		}()

		_ = os.Setenv("DOCKER_HOST", "example.com")

		c, err := StartCustom(testImage, DefaultTCP(80))
		require.True(t, errors.Is(err, ErrEnvClient))
		require.Nil(t, c)
	})

	t.Run("hostAddr returns docker host address", func(t *testing.T) {
		currentHost := os.Getenv("DOCKER_HOST")

		defer func() {
			_ = os.Setenv("DOCKER_HOST", currentHost)
		}()

		_ = os.Setenv("DOCKER_HOST", "tcp://1.1.1.1:2375")

		d := &docker{}
		addr := d.hostAddr()
		require.Equal(t, "1.1.1.1", addr)
	})

	t.Run("hostAddr falls back to local", func(t *testing.T) {
		t.Run("wrong url", func(t *testing.T) {
			currentHost := os.Getenv("DOCKER_HOST")

			defer func() {
				_ = os.Setenv("DOCKER_HOST", currentHost)
			}()

			_ = os.Setenv("DOCKER_HOST", ":")

			d := &docker{}
			addr := d.hostAddr()
			require.Equal(t, localhostAddr, addr)
		})

		t.Run("unix socket", func(t *testing.T) {
			currentHost := os.Getenv("DOCKER_HOST")

			defer func() {
				_ = os.Setenv("DOCKER_HOST", currentHost)
			}()

			_ = os.Setenv("DOCKER_HOST", "unix:///run/user/1000/docker.sock")

			d := &docker{}
			addr := d.hostAddr()
			require.Equal(t, localhostAddr, addr)
		})
	})
}
