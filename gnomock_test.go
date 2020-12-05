package gnomock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/stretchr/testify/require"
)

const testImage = "docker.io/orlangure/gnomock-test-image"
const goodPort80 = 80
const goodPort8080 = 8080
const badPort = 8000

func TestGnomock_happyFlow(t *testing.T) {
	t.Parallel()

	namedPorts := gnomock.NamedPorts{
		"web80":   gnomock.TCP(goodPort80),
		"web8080": gnomock.TCP(goodPort8080),
	}
	container, err := gnomock.StartCustom(
		testImage, namedPorts,
		gnomock.WithHealthCheckInterval(time.Microsecond*500),
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithInit(initf),
		gnomock.WithContext(context.Background()),
		gnomock.WithTimeout(time.Second*30),
		gnomock.WithEnv("GNOMOCK_TEST_1=foo"),
		gnomock.WithEnv("GNOMOCK_TEST_2=bar"),
	)

	require.NoError(t, err)
	require.NotNil(t, container)

	addr := fmt.Sprintf("http://%s/", container.Address("web80"))
	requireResponse(t, addr, "80")

	addr = fmt.Sprintf("http://%s/", container.Address("web8080"))
	requireResponse(t, addr, "8080")

	t.Run("default address is empty when no default port set", func(t *testing.T) {
		require.Empty(t, container.DefaultAddress())
	})

	t.Run("wrong port not found", func(t *testing.T) {
		_, err := container.Ports.Find("tcp", 1234)
		require.True(t, errors.Is(err, gnomock.ErrPortNotFound))
	})

	t.Run("default port is zero when no default port set", func(t *testing.T) {
		require.Zero(t, container.DefaultPort())
	})

	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_wrongPort(t *testing.T) {
	t.Parallel()

	container, err := gnomock.StartCustom(
		testImage, gnomock.DefaultTCP(badPort),
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithTimeout(time.Millisecond*50),
	)
	require.Error(t, err)
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_cancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	container, err := gnomock.StartCustom(
		testImage, gnomock.DefaultTCP(badPort),
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithContext(ctx),
	)
	require.True(t, errors.Is(err, context.Canceled))
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_defaultHealthcheck(t *testing.T) {
	t.Parallel()

	container, err := gnomock.StartCustom(testImage, gnomock.DefaultTCP(badPort))

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	// there is no error since healthcheck never returns an error
	require.NoError(t, err)
}

func TestGnomock_initError(t *testing.T) {
	t.Parallel()

	errNope := fmt.Errorf("nope")
	initWithErr := func(context.Context, *gnomock.Container) error {
		return errNope
	}

	container, err := gnomock.StartCustom(
		testImage, gnomock.DefaultTCP(goodPort80),
		gnomock.WithInit(initWithErr),
	)
	require.True(t, errors.Is(err, errNope))
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_cantStart(t *testing.T) {
	t.Parallel()

	container, err := gnomock.StartCustom(
		"docker.io/orlangure/noimage",
		gnomock.DefaultTCP(goodPort80),
	)
	require.Error(t, err)
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_withLogWriter(t *testing.T) {
	t.Parallel()

	r, w := io.Pipe()

	container, err := gnomock.StartCustom(
		testImage, gnomock.DefaultTCP(goodPort80),
		gnomock.WithLogWriter(w),
	)
	require.NoError(t, err)

	signal := make(chan struct{})

	go func() {
		defer close(signal)

		log, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, "starting with env1 = '', env2 = ''\n", string(log))
	}()

	require.NoError(t, gnomock.Stop(container))

	require.NoError(t, w.Close())
	<-signal
	require.NoError(t, r.Close())
}

func healthcheck(ctx context.Context, c *gnomock.Container) error {
	err := callRoot(fmt.Sprintf("http://%s/", c.Address("web80")))
	if err != nil {
		return err
	}

	err = callRoot(fmt.Sprintf("http://%s/", c.Address("web8080")))
	if err != nil {
		return err
	}

	return nil
}

func callRoot(addr string) error {
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

func initf(context.Context, *gnomock.Container) error {
	return nil
}

func requireResponse(t *testing.T, url string, expected string) {
	resp, err := http.Get(url)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)

	require.NoError(t, err)
	require.Equal(t, expected, string(body))
}
