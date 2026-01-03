package jaeger_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/jaeger"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	p := jaeger.Preset(jaeger.WithVersion("1.57"))
	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	addr := container.Address("web")
	url := fmt.Sprintf("http://%s/", addr)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
