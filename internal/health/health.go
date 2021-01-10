// Package health includes common health check functions to use in other
// packages.
package health

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// HTTPGet returns no error when a request to the provided address (host:port)
// completes with a status < 400 (Bad Request).
func HTTPGet(ctx context.Context, addr string) error {
	if !strings.HasPrefix(addr, "http://") {
		addr = fmt.Sprintf("http://%s", addr)
	}

	r, err := http.NewRequest(http.MethodGet, addr, nil) // nolint:gosec
	if err != nil {
		return err
	}

	r = r.WithContext(ctx)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected response code: %d", res.StatusCode)
	}

	return nil
}
