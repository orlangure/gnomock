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

func TestGnomock_happyFlow(t *testing.T) {
	t.Parallel()

	opts := gnomock.Options{
		Image:               "docker.io/jwilder/whoami",
		Version:             "latest",
		Port:                8000,
		HealthcheckInterval: time.Microsecond * 500,
		Healthcheck:         healthcheck,
		Init:                initf,
		Ctx:                 context.Background(),
		StartTimeout:        time.Second * 10,
		WaitTimeout:         time.Second * 1,
	}

	container, err := gnomock.Start(opts)

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

	opts := gnomock.Options{
		Image:       "docker.io/jwilder/whoami",
		Port:        80,
		Healthcheck: healthcheck,
		WaitTimeout: time.Millisecond * 50,
	}

	container, err := gnomock.Start(opts)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.Error(t, err)
	require.NotNil(t, container)
}

func TestGnomock_cancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	opts := gnomock.Options{
		Image:       "docker.io/jwilder/whoami",
		Port:        80,
		Healthcheck: healthcheck,
		Ctx:         ctx,
	}

	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	container, err := gnomock.Start(opts)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.True(t, errors.Is(err, context.Canceled))
}

func TestGnomock_defaultHealthcheck(t *testing.T) {
	t.Parallel()

	opts := gnomock.Options{
		Image: "docker.io/jwilder/whoami",
		Port:  81,
	}

	container, err := gnomock.Start(opts)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	// there is no error since healthcheck never returns an error
	require.NoError(t, err)
}

func TestGnomock_requiredOptions_port(t *testing.T) {
	t.Parallel()

	opts := gnomock.Options{
		Image: "docker.io/jwilder/whoami",
	}

	container, err := gnomock.Start(opts)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.True(t, errors.Is(err, gnomock.ErrPortNotSet))
}

func TestGnomock_requiredOptions_image(t *testing.T) {
	t.Parallel()

	opts := gnomock.Options{
		Port: 8000,
	}

	container, err := gnomock.Start(opts)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.True(t, errors.Is(err, gnomock.ErrImageNotSet))
}

func TestGnomock_initError(t *testing.T) {
	t.Parallel()

	errNope := fmt.Errorf("nope")
	opts := gnomock.Options{
		Image: "docker.io/jwilder/whoami",
		Port:  8000,
		Init: func(*gnomock.Container) error {
			return errNope
		},
	}

	container, err := gnomock.Start(opts)

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
