package gnomock

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type docker struct {
	client *client.Client
}

func dockerConnect() (*docker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEnvClient, err)
	}

	return &docker{cli}, nil
}

func (d *docker) pullImage(ctx context.Context, image string) error {
	reader, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("can't pull image: %w", err)
	}

	defer func() {
		closeErr := reader.Close()

		if err == nil {
			err = closeErr
		}
	}()

	_, err = ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("can't read server output: %w", err)
	}

	return nil
}

func (d *docker) startContainer(ctx context.Context, image string, port int) (*Container, error) {
	containerPort := nat.Port(fmt.Sprintf("%d/tcp", port))
	containerConfig := &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
	}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{
				{
					HostIP: "127.0.0.1",
				},
			},
		},
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return nil, fmt.Errorf("can't create container: %w", err)
	}

	err = d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("can't start container %s: %w", resp.ID, err)
	}

	json, err := d.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("can't inspect container %s: %w", resp.ID, err)
	}

	boundPorts, found := json.NetworkSettings.Ports[containerPort]
	if !found || len(boundPorts) == 0 {
		return nil, fmt.Errorf("can't bind port for %s", containerPort)
	}

	boundPort := boundPorts[0]
	container := &Container{
		ID:   json.ID,
		Host: boundPort.HostIP,
		Port: boundPort.HostPort,
	}

	return container, nil
}

func (d *docker) stopContainer(ctx context.Context, id string) error {
	err := d.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("can't stop container %s: %w", id, err)
	}

	return nil
}
