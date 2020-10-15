package gnomockd_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/mongo" // this is only to prevent error 404
	"github.com/stretchr/testify/require"
)

//nolint:bodyclose
func TestGnomockd(t *testing.T) {
	t.Run("start with preset not found", func(t *testing.T) {
		t.Parallel()

		h := gnomockd.Handler()
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/foobar", nil)
		h.ServeHTTP(w, r)

		res := w.Result()

		defer func() { require.NoError(t, res.Body.Close()) }()

		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("start with empty body", func(t *testing.T) {
		t.Parallel()

		h := gnomockd.Handler()
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/mongo", nil)
		h.ServeHTTP(w, r)

		res := w.Result()

		defer func() { require.NoError(t, res.Body.Close()) }()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("stop with empty body", func(t *testing.T) {
		t.Parallel()

		h := gnomockd.Handler()
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", nil)
		h.ServeHTTP(w, r)

		res := w.Result()

		defer func() { require.NoError(t, res.Body.Close()) }()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("stop with no ID", func(t *testing.T) {
		t.Parallel()

		h := gnomockd.Handler()
		buf := bytes.NewBuffer([]byte("{}"))
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
		h.ServeHTTP(w, r)

		res := w.Result()

		defer func() { require.NoError(t, res.Body.Close()) }()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
