package errors_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/errors"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPresetNotFoundError(t *testing.T) {
	err := errors.NewPresetNotFoundError("invalid")
	require.Equal(t, "preset 'invalid' not found", err.Error())
	require.Equal(t, http.StatusNotFound, errors.ErrorCode(err))
}

func TestInvalidStartRequestError(t *testing.T) {
	rootErr := fmt.Errorf("bad input")
	err := errors.NewInvalidStartRequestError(rootErr)
	require.Equal(t, "invalid start request: bad input", err.Error())
	require.Equal(t, http.StatusBadRequest, errors.ErrorCode(err))
}

func TestStartFailedError(t *testing.T) {
	c := &gnomock.Container{ID: "foobar"}
	rootErr := fmt.Errorf("bad container")
	err := errors.NewStartFailedError(rootErr, c)
	require.Equal(t, "start failed: bad container", err.Error())
	require.Equal(t, http.StatusInternalServerError, errors.ErrorCode(err))
}

func TestInvalidStopRequestError(t *testing.T) {
	rootErr := fmt.Errorf("bad input")
	err := errors.InvalidStopRequestError(rootErr)
	require.Equal(t, "invalid stop request: bad input", err.Error())
	require.Equal(t, http.StatusBadRequest, errors.ErrorCode(err))
}

func TestStopFailedError(t *testing.T) {
	c := &gnomock.Container{ID: "foobar"}
	rootErr := fmt.Errorf("bad container")
	err := errors.StopFailedError(rootErr, c)
	require.Equal(t, "stop failed: bad container", err.Error())
	require.Equal(t, http.StatusInternalServerError, errors.ErrorCode(err))
}
