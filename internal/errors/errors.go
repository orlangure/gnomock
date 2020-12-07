// Package errors exposes errors types used by gnomockd endpoints.
package errors

import (
	"fmt"
	"net/http"

	"github.com/orlangure/gnomock"
)

// NewPresetNotFoundError is returned when an invalid/unknown preset name was
// used.
func NewPresetNotFoundError(name string) error {
	return presetNotFoundError{
		name:   name,
		ErrStr: fmt.Sprintf("preset '%s' not found", name),
	}
}

type presetNotFoundError struct {
	name   string
	ErrStr string `json:"error"`
}

func (e presetNotFoundError) Error() string {
	return e.ErrStr
}

// NewInvalidStartRequestError means that the request parameters of /start call
// were invalid
func NewInvalidStartRequestError(err error) error {
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

// NewStartFailedError means that the container failed to start for some reason
func NewStartFailedError(err error, c *gnomock.Container) error {
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

// ErrorCode returns HTTP response code for the provided error
func ErrorCode(err error) int {
	switch err.(type) {
	case invalidStartRequestError, invalidStopRequestError:
		return http.StatusBadRequest
	case presetNotFoundError:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
