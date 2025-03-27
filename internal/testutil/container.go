// Package testutil includes utilities used in test code of other packages.
package testutil

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// ListContainerByID returns a list of Containers for the given container id.
func ListContainerByID(cli *client.Client, id string) ([]container.Summary, error) {
	return cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "id",
			Value: id,
		}),
	})
}
