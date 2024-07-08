package gnomock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestGnomock_happyFlow(t *testing.T) {
	t.Parallel()

	namedPorts := gnomock.NamedPorts{
		"web80":   gnomock.TCP(testutil.GoodPort80),
		"web8080": gnomock.TCP(testutil.GoodPort8080),
	}
	container, err := gnomock.StartCustom(
		testutil.TestImage, namedPorts,
		gnomock.WithHealthCheckInterval(time.Microsecond*500),
		gnomock.WithHealthCheck(testutil.Healthcheck),
		gnomock.WithInit(initf),
		gnomock.WithContext(context.Background()),
		gnomock.WithTimeout(time.Minute),
		gnomock.WithEnv("GNOMOCK_TEST_1=foo"),
		gnomock.WithEnv("GNOMOCK_TEST_2=bar"),
		gnomock.WithRegistryAuth(""),
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
		testutil.TestImage, gnomock.DefaultTCP(testutil.BadPort),
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
		testutil.TestImage, gnomock.DefaultTCP(testutil.BadPort),
		gnomock.WithContext(ctx),
	)
	require.True(t, errors.Is(err, context.Canceled))
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_customHealthcheck(t *testing.T) {
	t.Parallel()

	container, err := gnomock.StartCustom(
		testutil.TestImage, gnomock.DefaultTCP(testutil.BadPort),
		gnomock.WithTimeout(time.Second*10),
		gnomock.WithHealthCheck(failingHealthcheck),
	)

	defer func() {
		require.NoError(t, gnomock.Stop(container))
	}()

	require.Error(t, err)
}

func TestGnomock_initError(t *testing.T) {
	t.Parallel()

	errNope := fmt.Errorf("nope")
	initWithErr := func(context.Context, *gnomock.Container) error {
		return errNope
	}

	container, err := gnomock.StartCustom(
		testutil.TestImage, gnomock.DefaultTCP(testutil.GoodPort80),
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
		gnomock.DefaultTCP(testutil.GoodPort80),
	)
	require.Error(t, err)
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_withDebugMode(t *testing.T) {
	t.Parallel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	container, err := gnomock.StartCustom(
		testutil.TestImage, gnomock.DefaultTCP(testutil.GoodPort80),
	)
	require.NoError(t, err)
	require.NotNil(t, container)
	require.NoError(t, gnomock.Stop(container))

	containerList, err := testutil.ListContainerByID(cli, container.ID)
	require.NoError(t, err)
	require.Len(t, containerList, 0)

	container, err = gnomock.StartCustom(
		testutil.TestImage, gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithDebugMode(),
	)
	require.NoError(t, err)
	require.NotNil(t, container)

	containerList, err = testutil.ListContainerByID(cli, container.ID)
	require.NoError(t, err)
	require.Len(t, containerList, 1)
	require.NoError(t, gnomock.Stop(container))

	containerList, err = testutil.ListContainerByID(cli, container.ID)
	require.NoError(t, err)
	require.Len(t, containerList, 0)
	require.NoError(t, cli.Close())
}

func TestGnomock_withLogWriter(t *testing.T) {
	t.Parallel()

	r, w := io.Pipe()

	container, err := gnomock.StartCustom(
		testutil.TestImage, gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithLogWriter(w),
	)
	require.NoError(t, err)

	signal := make(chan struct{})

	go func() {
		defer close(signal)

		log, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(log), "starting with env1 = '', env2 = ''\n")
	}()

	require.NoError(t, gnomock.Stop(container))

	require.NoError(t, w.Close())
	<-signal
	require.NoError(t, r.Close())
}

func TestGnomock_withCommand(t *testing.T) {
	t.Parallel()

	r, w := io.Pipe()

	container, err := gnomock.StartCustom(
		testutil.TestImage, gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithLogWriter(w),
		gnomock.WithCommand("foo", "bar"),
	)
	require.NoError(t, err)

	signal := make(chan struct{})

	go func() {
		defer close(signal)

		log, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(log), "[foo bar]")
	}()

	require.NoError(t, gnomock.Stop(container))

	require.NoError(t, w.Close())
	<-signal
	require.NoError(t, r.Close())
}

func TestGnomock_withEntrypoint(t *testing.T) {
	t.Run("overwriting entrypoint with the same entrypoint as the original image", func(t *testing.T) {
		r, w := io.Pipe()

		container, err := gnomock.StartCustom(
			testutil.TestImage,
			gnomock.DefaultTCP(testutil.GoodPort80),
			gnomock.WithLogWriter(w),
			gnomock.WithEntrypoint("/app"),
			gnomock.WithCommand("foo", "bar"),
		)
		require.NoError(t, err)

		signal := make(chan struct{})

		go func() {
			defer close(signal)

			log, err := io.ReadAll(r)
			require.NoError(t, err)
			require.Contains(t, string(log), "[foo bar]")
		}()

		require.NoError(t, gnomock.Stop(container))

		require.NoError(t, w.Close())
		<-signal
		require.NoError(t, r.Close())
	})
	t.Run("overwriting entrypoint with a new entrypoint", func(t *testing.T) {
		_, err := gnomock.StartCustom(
			testutil.TestImage,
			gnomock.DefaultTCP(testutil.GoodPort80),
			gnomock.WithEntrypoint("echo"),
		)

		require.ErrorContains(t, err, "\"echo\": executable file not found in $PATH")
	})
}

// See https://github.com/orlangure/gnomock/issues/302
func TestGnomock_witUseLocalImagesFirst(t *testing.T) {
	t.Parallel()

	const (
		mongoImage         = "docker.io/library/mongo:4.4"
		circleciMongoImage = "docker.io/circleci/mongo:4.4"
	)

	// this block will ensure having a local library/mongo image
	container, err := gnomock.StartCustom(
		mongoImage,
		gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithUseLocalImagesFirst(),
	)
	require.NoError(t, err)
	require.NotNil(t, container)
	require.NoError(t, gnomock.Stop(container))

	t.Run("library mongo image", func(t *testing.T) {
		t.Parallel()

		// this actually uses the local image
		container, err := gnomock.StartCustom(
			mongoImage,
			gnomock.DefaultTCP(testutil.GoodPort80),
			gnomock.WithUseLocalImagesFirst(),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
		require.NoError(t, gnomock.Stop(container))
	})

	t.Run("circleci mongo image", func(t *testing.T) {
		t.Parallel()

		container, err := gnomock.StartCustom(
			circleciMongoImage,
			gnomock.DefaultTCP(testutil.GoodPort80),
			gnomock.WithUseLocalImagesFirst(),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
		require.NoError(t, gnomock.Stop(container))
	})

	t.Run("local image", func(t *testing.T) {
		t.Parallel()
		t.Skip("Enable this test when building from Dockerfile is supported")

		container, err := gnomock.StartCustom(
			"local-image",
			gnomock.DefaultTCP(testutil.GoodPort80),
			gnomock.WithUseLocalImagesFirst(),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
		require.NoError(t, gnomock.Stop(container))
	})
}

func TestGnomock_withExtraHosts(t *testing.T) {
	t.Parallel()

	const (
		busyboxImage = "docker.io/library/busybox:1.35.0"
		retries      = "5"
	)

	container, err := gnomock.StartCustom(
		busyboxImage,
		gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithExtraHosts([]string{"test:127.0.0.1"}),
		gnomock.WithCommand("ping", "-c", retries, "test"),
	)
	require.NoError(t, err)
	require.NotNil(t, container)
	require.NoError(t, gnomock.Stop(container))

	container, err = gnomock.StartCustom(
		busyboxImage,
		gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithCommand("ping", "-c", retries, "test"),
	)
	require.Error(t, err)
	require.Nil(t, container)
	require.NoError(t, gnomock.Stop(container))

	container, err = gnomock.StartCustom(
		busyboxImage,
		gnomock.DefaultTCP(testutil.GoodPort80),
		gnomock.WithExtraHosts([]string{"test:127.0.0.1"}),
		gnomock.WithCommand("ping", "-c", retries, "google.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, container)
	require.NoError(t, gnomock.Stop(container))
}

func TestGnomock_withCustomImage(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: "docker.io/orlangure/noimage"}
	container, err := gnomock.Start(p, gnomock.WithCustomImage(testutil.TestImage))
	require.NoError(t, err)
	require.NotNil(t, container)
	require.NoError(t, gnomock.Stop(container))
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

	body, err := io.ReadAll(resp.Body)

	require.NoError(t, err)
	require.Equal(t, expected, string(body))
}

func failingHealthcheck(_ context.Context, _ *gnomock.Container) error {
	return fmt.Errorf("this container should not start")
}
