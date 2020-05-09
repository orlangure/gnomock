package gnomockd_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlangure/gnomockd/gnomockd"
	"github.com/stretchr/testify/require"
)

//nolint:bodyclose
func TestStart_notFound(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/foobar", nil)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

//nolint:bodyclose
func TestStart_emptyBody(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/mongo", nil)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

//nolint:bodyclose
func TestStop_emptyBody(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", nil)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

//nolint:bodyclose
func TestStop_noID(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	buf := bytes.NewBuffer([]byte("{}"))
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}
