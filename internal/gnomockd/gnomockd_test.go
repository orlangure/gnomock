package gnomockd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/mongo" // this is only to prevent error 404
	"github.com/stretchr/testify/require"
)

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

	t.Run("stop with wrong ID", func(t *testing.T) {
		t.Parallel()

		h := gnomockd.Handler()
		buf := bytes.NewBufferString(`{"id":"invalid"}`)
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
		h.ServeHTTP(w, r)

		res := w.Result()

		defer func() { require.NoError(t, res.Body.Close()) }()

		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("fixed host port using custom named ports", func(t *testing.T) {
		t.Parallel()

		port := gnomock.TCP(27017)
		port.HostPort = 43210

		body, err := json.Marshal(struct {
			Options gnomock.Options `json:"options"`
		}{
			gnomock.Options{
				CustomNamedPorts: gnomock.NamedPorts{
					gnomock.DefaultPort: port,
				},
			},
		})
		require.NoError(t, err)

		h := gnomockd.Handler()
		w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/mongo", bytes.NewBuffer(body))
		h.ServeHTTP(w, r)

		res := w.Result()

		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })
		require.Equal(t, http.StatusOK, res.StatusCode)

		body, err = io.ReadAll(res.Body)
		require.NoError(t, err)

		c := gnomock.Container{}
		require.NoError(t, json.Unmarshal(body, &c))
		require.Equal(t, 43210, c.DefaultPort())
	})
}
