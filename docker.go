package gnomock

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/orlangure/gnomock/internal/cleaner"
	"github.com/orlangure/gnomock/internal/health"
	"go.uber.org/zap"
)

const (
	localhostAddr             = "127.0.0.1"
	defaultStopTimeout        = time.Second * 1
	duplicateContainerPattern = `Conflict. The container name "(?:.+?)" is already in use by container "(\w+)". You have to remove \(or rename\) that container to be able to reuse that name.` // nolint:lll
	dockerSockAddr            = "/var/run/docker.sock"
)

var duplicateContainerRegexp = regexp.MustCompile(duplicateContainerPattern)

type docker struct {
	client *client.Client
	log    *zap.SugaredLogger

	// This lock is used to protect docker client from concurrent connections
	// with version negotiation. As of this moment, there is a data race in
	// docker client when version negotiation is requested. This data race is
	// not dangerous, but it triggers race detector alarms, so it should be
	// avoided. Currently the client still has this issue, so this is an
	// attempt to fix it locally by preventing concurrent connection using the
	// same client (mostly when `Stop` is called with multiple containers).
	//
	// https://github.com/moby/moby/pull/42379
	lock sync.Mutex
}

func (g *g) dockerConnect() (*docker, error) {
	g.log.Info("connecting to docker engine")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEnvClient, err)
	}

	g.log.Info("connected to docker engine")

	return &docker{client: cli, log: g.log}, nil
}

func (d *docker) isExistingLocalImage(ctx context.Context, image string) (bool, error) {
	images, err := d.client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("can't list image: %w", err)
	}

	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			if !strings.Contains(repoTag, "/") {
				repoTag = "library/" + repoTag
			}

			if strings.HasSuffix(image, repoTag) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (d *docker) pullImage(ctx context.Context, image string, cfg *Options) error {
	d.log.Info("pulling image")

	reader, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{
		RegistryAuth: cfg.Auth,
	})
	if err != nil {
		return fmt.Errorf("can't pull image: %w", err)
	}

	defer func() {
		closeErr := reader.Close()

		if err == nil {
			err = closeErr
		}
	}()

	_, err = io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("can't read server output: %w", err)
	}

	d.log.Info("image pulled")

	return nil
}

func (d *docker) startContainer(ctx context.Context, image string, ports NamedPorts, cfg *Options) (*Container, error) {
	d.log.Info("starting container")

	resp, err := d.prepareContainer(ctx, image, ports, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't prepare container: %w", err)
	}

	sidecarChan := make(chan string)

	go func() {
		defer close(sidecarChan)

		if cfg.DisableAutoCleanup || cfg.Debug {
			return
		}

		opts := []Option{
			WithDisableAutoCleanup(),
			WithHostMounts(dockerSockAddr, dockerSockAddr),
			WithHealthCheck(func(ctx context.Context, c *Container) error {
				return health.HTTPGet(ctx, c.DefaultAddress())
			}),
			WithInit(func(ctx context.Context, c *Container) error {
				return cleaner.Notify(context.Background(), c.DefaultAddress(), resp.ID)
			}),
		}
		if cfg.UseLocalImagesFirst {
			opts = append(opts, WithUseLocalImagesFirst())
		}

		if sc, err := StartCustom(
			cleaner.Image, DefaultTCP(cleaner.Port),
			opts...,
		); err == nil {
			sidecarChan <- sc.ID
		}
	}()

	err = d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("can't start container %s: %w", resp.ID, err)
	}

	container, err := d.waitForContainerNetwork(ctx, resp.ID, ports)
	if err != nil {
		return nil, fmt.Errorf("container network isn't ready: %w", err)
	}

	if sidecar, ok := <-sidecarChan; ok {
		container.ID = generateID(container.ID, sidecar)
	}

	d.log.Infow("container started", "container", container)

	return container, nil
}

func (d *docker) prepareContainer(
	ctx context.Context,
	image string,
	ports NamedPorts,
	cfg *Options,
) (*container.ContainerCreateCreatedBody, error) {
	pullImage := true

	if cfg.UseLocalImagesFirst {
		isExisting, err := d.isExistingLocalImage(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("can't list image: %w", err)
		}

		if isExisting {
			pullImage = false
		}
	}

	if pullImage {
		if err := d.pullImage(ctx, image, cfg); err != nil {
			return nil, fmt.Errorf("can't pull image: %w", err)
		}
	}

	resp, err := d.createContainer(ctx, image, ports, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't create container: %w", err)
	}

	return resp, err
}

func (d *docker) waitForContainerNetwork(ctx context.Context, id string, ports NamedPorts) (*Container, error) {
	d.log.Infow("waiting for container network", "container", id)

	tick := time.NewTicker(time.Millisecond * 250)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("container network is unavailable after timeout")
		case <-tick.C:
			containerJSON, err := d.client.ContainerInspect(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("can't inspect container %s: %w", id, err)
			}

			boundNamedPorts, err := d.boundNamedPorts(containerJSON, ports)
			if err != nil {
				return nil, fmt.Errorf("can't find bound ports: %w", err)
			}

			d.log.Infow("waiting for port allocation", "container", id)

			if len(boundNamedPorts) == len(ports) {
				return &Container{
					ID:      id,
					Host:    d.hostAddr(),
					Ports:   boundNamedPorts,
					gateway: containerJSON.NetworkSettings.Gateway,
				}, nil
			}
		}
	}
}

func (d *docker) exposedPorts(namedPorts NamedPorts) nat.PortSet {
	exposedPorts := make(nat.PortSet)

	for _, port := range namedPorts {
		containerPort := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		exposedPorts[nat.Port(containerPort)] = struct{}{}
	}

	return exposedPorts
}

func (d *docker) portBindings(exposedPorts nat.PortSet, ports NamedPorts) nat.PortMap {
	portBindings := make(nat.PortMap)

	// for the container to be accessible from another container, it cannot
	// listen on 127.0.0.1 as it will be accessed by gateway address (e.g
	// 172.17.0.1), so its port should be exposed everywhere
	hostAddr := d.hostAddr()
	if isInDocker() {
		hostAddr = "0.0.0.0"
	}

	for port := range exposedPorts {
		binding := nat.PortBinding{
			HostIP: hostAddr,
		}

		if pName, err := ports.Find(port.Proto(), port.Int()); err == nil {
			namedPort := ports.Get(pName)
			if namedPort.HostPort > 0 {
				binding.HostPort = strconv.Itoa(namedPort.HostPort)
			}
		}

		portBindings[port] = []nat.PortBinding{binding}
	}

	return portBindings
}

func (d *docker) createContainer(ctx context.Context, image string, ports NamedPorts, cfg *Options) (*container.ContainerCreateCreatedBody, error) { // nolint:lll
	exposedPorts := d.exposedPorts(ports)
	containerConfig := &container.Config{
		Image:        image,
		ExposedPorts: exposedPorts,
		Env:          cfg.Env,
	}

	if len(cfg.Cmd) > 0 {
		containerConfig.Cmd = cfg.Cmd
	}

	mounts := []mount.Mount{}
	for src, dst := range cfg.HostMounts {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: src,
			Target: dst,
		})
	}

	portBindings := d.portBindings(exposedPorts, ports)
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		AutoRemove:   true,
		Privileged:   cfg.Privileged,
		Mounts:       mounts,
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, cfg.ContainerName)
	if err == nil {
		return &resp, nil
	}

	matches := duplicateContainerRegexp.FindStringSubmatch(err.Error())
	if len(matches) == 2 {
		d.log.Infow("duplicate container found, stopping", "container", matches[1])

		err = d.client.ContainerRemove(ctx, matches[1], types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return nil, fmt.Errorf("can't remove existing container: %w", err)
		}

		resp, err = d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, cfg.ContainerName)
	}

	return &resp, err
}

func (d *docker) boundNamedPorts(json types.ContainerJSON, namedPorts NamedPorts) (NamedPorts, error) {
	boundNamedPorts := make(NamedPorts)

	for containerPort, bindings := range json.NetworkSettings.Ports {
		if len(bindings) == 0 {
			continue
		}

		hostPortNum, err := strconv.Atoi(bindings[0].HostPort)
		if err != nil {
			return nil, fmt.Errorf("invalid host port value '%s': %w", bindings[0].HostPort, err)
		}

		proto, intPort := containerPort.Proto(), containerPort.Int()

		portName, err := namedPorts.Find(proto, intPort)
		if err != nil {
			return nil, fmt.Errorf("can't find port %s/%d: %w", proto, intPort, err)
		}

		boundNamedPorts[portName] = Port{
			Protocol: proto,
			Port:     hostPortNum,
		}
	}

	return boundNamedPorts, nil
}

func (d *docker) readLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	d.log.Info("starting container logs forwarder")

	logsOptions := types.ContainerLogsOptions{
		ShowStderr: true, ShowStdout: true, Follow: true,
	}

	rc, err := d.client.ContainerLogs(ctx, id, logsOptions)
	if err != nil {
		return nil, fmt.Errorf("can't read logs: %w", err)
	}

	d.log.Info("container logs forwarder ready")

	return rc, nil
}

func (d *docker) stopContainer(ctx context.Context, id string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	stopTimeout := defaultStopTimeout

	err := d.client.ContainerStop(ctx, id, &stopTimeout)
	if err != nil {
		return fmt.Errorf("can't stop container %s: %w", id, err)
	}

	return nil
}

// hostAddr returns an address of a host that runs the containers. If
// DOCKER_HOST environment variable is not set, if its value is an invalid URL,
// or if it is a `unix:///` socket address, it returns local address.
func (d *docker) hostAddr() string {
	if dh := os.Getenv("DOCKER_HOST"); dh != "" {
		u, err := url.Parse(dh)
		if err == nil {
			if host := u.Hostname(); host != "" {
				return host
			}
		}
	}

	return localhostAddr
}
