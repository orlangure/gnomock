package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlangure/gnomock/internal/health"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestHTTPGet(t *testing.T) {
	ctx := context.Background()

	t.Run("wrong url", func(t *testing.T) {
		require.Error(t, health.HTTPGet(ctx, "foo::bar"))
	})

	t.Run("addr unavailable", func(t *testing.T) {
		require.Error(t, health.HTTPGet(ctx, "random.local.host"))
	})

	t.Run("invalid status code", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer s.Close()

		err := health.HTTPGet(ctx, s.URL)
		require.EqualError(t, err, "unexpected response code: 500")
	})

	t.Run("success", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()

		require.NoError(t, health.HTTPGet(ctx, s.URL))
	})
}
