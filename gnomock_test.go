//nolint:gosec
package gnomock_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/stretchr/testify/require"
)

const testImage = "docker.io/jwilder/whoami"
const testPort = 8000

func TestGnomock_happyFlow(t *testing.T) {
	t.Parallel()

	container, err := gnomock.Start(
		testImage, testPort,
		gnomock.WithHealthCheckInterval(time.Microsecond*500),
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithInit(initf),
		gnomock.WithContext(context.Background()),
		gnomock.WithStartTimeout(time.Second*10),
		gnomock.WithWaitTimeout(time.Second*1),
	)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.NoError(t, err)
	require.NotNil(t, container)

	addr := fmt.Sprintf("http://%s/", container.Address())
	resp, err := http.Get(addr)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGnomock_wrongPort(t *testing.T) {
	t.Parallel()

	container, err := gnomock.Start(
		testImage, 80,
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithWaitTimeout(time.Millisecond*50),
	)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.Error(t, err)
	require.NotNil(t, container)
}

func TestGnomock_cancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	container, err := gnomock.Start(
		testImage, 80,
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithContext(ctx),
	)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.True(t, errors.Is(err, context.Canceled))
}

func TestGnomock_defaultHealthcheck(t *testing.T) {
	t.Parallel()

	container, err := gnomock.Start(testImage, 81)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	// there is no error since healthcheck never returns an error
	require.NoError(t, err)
}

func TestGnomock_initError(t *testing.T) {
	t.Parallel()

	errNope := fmt.Errorf("nope")
	initWithErr := func(*gnomock.Container) error {
		return errNope
	}

	container, err := gnomock.Start(
		testImage, testPort,
		gnomock.WithInit(initWithErr),
	)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.True(t, errors.Is(err, errNope))
}

func healthcheck(host, port string) (err error) {
	addr := fmt.Sprintf("http://%s:%s/", host, port)

	resp, err := http.Get(addr)
	if err != nil {
		return fmt.Errorf("can't GET %s: %w", addr, err)
	}

	defer func() {
		closeErr := resp.Body.Close()

		if err == nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

func initf(*gnomock.Container) error {
	return nil
}
