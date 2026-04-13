// Package testutil includes utilities used in test code of other packages.
package testutil

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// ListContainerByID returns a list of Containers for the given container id.
func ListContainerByID(cli *client.Client, id string) ([]container.Summary, error) {
	result, err := cli.ContainerList(context.Background(), client.ContainerListOptions{
		All:     true,
		Filters: make(client.Filters).Add("id", id),
	})
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
