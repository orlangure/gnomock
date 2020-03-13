package gnomock

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const localhostAddr = "127.0.0.1"

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

func (d *docker) startContainer(ctx context.Context, image string, namedPorts NamedPorts) (*Container, error) {
	exposedPorts := d.exposedPorts(namedPorts)
	containerConfig := &container.Config{
		Image:        image,
		ExposedPorts: exposedPorts,
	}
	portBindings := d.portBindings(exposedPorts)
	hostConfig := &container.HostConfig{PortBindings: portBindings}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return nil, fmt.Errorf("can't create container: %w", err)
	}

	err = d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("can't start container %s: %w", resp.ID, err)
	}

	containerJSON, err := d.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("can't inspect container %s: %w", resp.ID, err)
	}

	boundNamedPorts, err := d.boundNamedPorts(containerJSON, namedPorts)
	if err != nil {
		return nil, fmt.Errorf("can't find bound ports: %w", err)
	}

	container := &Container{
		ID:    containerJSON.ID,
		Host:  localhostAddr,
		Ports: boundNamedPorts,
	}

	return container, nil
}

func (d *docker) exposedPorts(namedPorts NamedPorts) nat.PortSet {
	exposedPorts := make(nat.PortSet)

	for _, port := range namedPorts {
		containerPort := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		exposedPorts[nat.Port(containerPort)] = struct{}{}
	}

	return exposedPorts
}

func (d *docker) portBindings(exposedPorts nat.PortSet) nat.PortMap {
	portBindings := make(nat.PortMap)

	for port := range exposedPorts {
		portBindings[port] = []nat.PortBinding{
			{
				HostIP: localhostAddr,
			},
		}
	}

	return portBindings
}

func (d *docker) boundNamedPorts(json types.ContainerJSON, namedPorts NamedPorts) (NamedPorts, error) {
	boundNamedPorts := make(NamedPorts)

	for containerPort, bindings := range json.NetworkSettings.Ports {
		if len(bindings) == 0 {
			continue
		}

		hostPortNum, err := strconv.Atoi(bindings[0].HostPort)
		if err != nil {
			return nil, err
		}

		portName, err := namedPorts.Find(containerPort.Proto(), containerPort.Int())
		if err != nil {
			return nil, err
		}

		boundNamedPorts[portName] = Port{containerPort.Proto(), hostPortNum}
	}

	return boundNamedPorts, nil
}

func (d *docker) stopContainer(ctx context.Context, id string) error {
	err := d.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("can't stop container %s: %w", id, err)
	}

	return nil
}
