package cleaner_test

import (
	"context"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/cleaner"
	"github.com/orlangure/gnomock/internal/health"
	"github.com/orlangure/gnomock/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestCleaner(t *testing.T) {
	t.Parallel()

	p := &testutil.TestPreset{Img: testutil.TestImage}
	targetContainer, err := gnomock.Start(p, gnomock.WithDisableAutoCleanup())
	require.NoError(t, err)

	rootCtx, cancel := context.WithCancel(context.Background())

	cleanerContainer, err := gnomock.StartCustom(
		cleaner.Image, gnomock.DefaultTCP(cleaner.Port),
		gnomock.WithDisableAutoCleanup(),
		gnomock.WithHostMounts("/var/run/docker.sock", "/var/run/docker.sock"),
		gnomock.WithHealthCheck(func(ctx context.Context, c *gnomock.Container) error {
			return health.HTTPGet(ctx, c.DefaultAddress())
		}),
		gnomock.WithInit(func(ctx context.Context, c *gnomock.Container) error {
			return cleaner.Notify(rootCtx, c.DefaultAddress(), targetContainer.ID)
		}),
	)
	require.NoError(t, err)

	time.Sleep(time.Second)
	cancel()
	time.Sleep(time.Second * 5)

	// both stop calls cause errors because both containers no longer exist
	require.Error(t, gnomock.Stop(targetContainer))
	require.Error(t, gnomock.Stop(cleanerContainer))
}

func TestCleaner_wrongRequest(t *testing.T) {
	t.Parallel()

	require.Error(t, cleaner.Notify(context.Background(), "invalid::url", ""))
}
