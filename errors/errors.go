// Package errors exposes errors types used by gnomockd endpoints
package errors

import (
	"fmt"
	"net/http"

	"github.com/orlangure/gnomock"
)

// InvalidStartRequestError means that the request parameters of /start call
// were invalid
func InvalidStartRequestError(err error) error {
	return invalidStartRequestError{
		err:    err,
		ErrStr: fmt.Sprintf("invalid start request: %v", err),
	}
}

type invalidStartRequestError struct {
	err    error
	ErrStr string `json:"error"`
}

func (e invalidStartRequestError) Error() string {
	return e.ErrStr
}

// StartFailedError means that the container failed to start for some reason
func StartFailedError(err error, c *gnomock.Container) error {
	return startFailedError{
		err:       err,
		ErrStr:    fmt.Sprintf("start failed: %v", err),
		Container: c,
	}
}

type startFailedError struct {
	err       error
	ErrStr    string             `json:"error"`
	Container *gnomock.Container `json:"container,omitempty"`
}

func (e startFailedError) Error() string {
	return e.ErrStr
}

// InvalidStopRequestError means that the request parameters of /stop call were
// invalid
func InvalidStopRequestError(err error) error {
	return invalidStopRequestError{
		err:    err,
		ErrStr: fmt.Sprintf("invalid stop request: %v", err),
	}
}

type invalidStopRequestError struct {
	err    error
	ErrStr string `json:"error"`
}

func (e invalidStopRequestError) Error() string {
	return e.ErrStr
}

// StopFailedError means that the container failed to stop
func StopFailedError(err error, c *gnomock.Container) error {
	return stopFailedError{
		err:       err,
		ErrStr:    fmt.Sprintf("stop failed: %v", err),
		Container: c,
	}
}

type stopFailedError struct {
	err       error
	ErrStr    string             `json:"error"`
	Container *gnomock.Container `json:"container,omitempty"`
}

func (e stopFailedError) Error() string {
	return e.ErrStr
}

// PrepareResponseError means that something happened during response
// preparation. The action could have succeeded, but there was a problem
// sending a response
func PrepareResponseError(err error, c *gnomock.Container) error {
	return prepareResponseError{
		err:       err,
		ErrStr:    fmt.Sprintf("prepare response failed: %v", err),
		Container: c,
	}
}

type prepareResponseError struct {
	err       error
	ErrStr    string             `json:"error"`
	Container *gnomock.Container `json:"container,omitempty"`
}

func (e prepareResponseError) Error() string {
	return e.ErrStr
}

// ErrorCode returns HTTP response code for the provided error
func ErrorCode(err error) int {
	switch err.(type) {
	case invalidStartRequestError, invalidStopRequestError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
