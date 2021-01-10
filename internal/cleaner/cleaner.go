// Package cleaner exposes an API to communicate with Gnomock Cleaner, a
// separate process that waits for incoming connections and terminates running
// containers when these connections are closed.
package cleaner

import (
	"context"
	"fmt"
	"net/http"
)

// Image and Port to create Cleaner containers.
const (
	Image = "docker.io/orlangure/gnomock-cleaner"
	Port  = 8008
)

// Notify sends a new request to the cleaner process running at the provided
// address. This connection hangs (cleaner does not respond) until the provided
// context is canceled. After that, the connection terminates, and it cleaner
// kill the container with the provided id.
func Notify(ctx context.Context, addr, id string) error {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/sync/%s", addr, id), nil)
	if err != nil {
		return err
	}

	go http.DefaultClient.Do(r.WithContext(ctx)) // nolint:errcheck

	return nil
}
