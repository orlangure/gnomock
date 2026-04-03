package gnomock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/orlangure/gnomock/internal/cleaner"
	"github.com/orlangure/gnomock/internal/health"
	"go.uber.org/zap"
)

const (
	localhostAddr             = "127.0.0.1"
	defaultStopTimeoutSec     = 1
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

	cli, err := client.New(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Join(ErrEnvClient, err)
	}

	g.log.Info("connected to docker engine")

	return &docker{client: cli, log: g.log}, nil
}

func (d *docker) isExistingLocalImage(ctx context.Context, image string) (bool, error) {
	result, err := d.client.ImageList(ctx, client.ImageListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("can't list image: %w", err)
	}

	for _, img := range result.Items {
		for _, repoTag := range img.RepoTags {
			if image == repoTag {
				return true, nil
			}

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

	resp, err := d.client.ImagePull(ctx, image, client.ImagePullOptions{
		RegistryAuth: cfg.Auth,
	})
	if err != nil {
		return fmt.Errorf("can't pull image: %w", err)
	}

	defer func() {
		closeErr := resp.Close()

		if err == nil {
			err = closeErr
		}
	}()

	_, err = io.ReadAll(resp)
	if err != nil {
		return fmt.Errorf("can't read server output: %w", err)
	}

	d.log.Info("image pulled")

	return nil
}

func (d *docker) startContainer(ctx context.Context, image string, ports NamedPorts, cfg *Options) (*Container, error) {
	if cfg.Reuse {
		container, ok, err := d.findReusableContainer(ctx, image, ports, cfg)
		if err != nil {
			return nil, err
		}

		if ok {
			d.log.Info("re-using container")
			return container, nil
		}
	}

	d.log.Info("starting container")

	resp, err := d.prepareContainer(ctx, image, ports, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't prepare container: %w", err)
	}

	sidecarChan, cleanupCancel := d.setupContainerCleanup(resp.ID, cfg)

	_, err = d.client.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{})
	if err != nil {
		cleanupCancel()
		return nil, fmt.Errorf("can't start container %s: %w", resp.ID, err)
	}

	container, err := d.waitForContainerNetwork(ctx, resp.ID, ports)
	if err != nil {
		cleanupCancel()
		return nil, fmt.Errorf("container network isn't ready: %w", err)
	}

	if sidecar, ok := <-sidecarChan; ok {
		container.ID = generateID(container.ID, sidecar)
	}

	d.log.Infow("container started", "container", container)

	return container, nil
}

func (d *docker) setupContainerCleanup(id string, cfg *Options) (chan string, context.CancelFunc) {
	sidecarChan := make(chan string, 1)
	bctx, bcancel := context.WithCancel(context.Background())

	go func() {
		defer close(sidecarChan)

		if cfg.DisableAutoCleanup || cfg.Reuse || cfg.Debug {
			return
		}

		opts := []Option{
			WithDisableAutoCleanup(),
			WithHostMounts(dockerSockAddr, dockerSockAddr),
			WithHealthCheck(func(ctx context.Context, c *Container) error {
				return health.HTTPGet(ctx, c.DefaultAddress())
			}),
			WithInit(func(_ context.Context, c *Container) error {
				return cleaner.Notify(bctx, c.DefaultAddress(), id)
			}),
			WithContext(bctx),
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

	return sidecarChan, bcancel
}

func (d *docker) prepareContainer(
	ctx context.Context,
	image string,
	ports NamedPorts,
	cfg *Options,
) (*client.ContainerCreateResult, error) {
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
			inspectResult, err := d.client.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
			if err != nil {
				return nil, fmt.Errorf("can't inspect container %s: %w", id, err)
			}

			boundNamedPorts, err := d.boundNamedPorts(inspectResult.Container, ports)
			if err != nil {
				return nil, fmt.Errorf("can't find bound ports: %w", err)
			}

			d.log.Infow("waiting for port allocation", "container", id)

			if len(boundNamedPorts) == len(ports) {
				var gateway string
				for _, ep := range inspectResult.Container.NetworkSettings.Networks {
					if ep != nil && ep.Gateway.IsValid() {
						gateway = ep.Gateway.String()
						break
					}
				}

				return &Container{
					ID:      id,
					Host:    d.hostAddr(),
					Ports:   boundNamedPorts,
					gateway: gateway,
				}, nil
			}
		}
	}
}

func (d *docker) exposedPorts(namedPorts NamedPorts) network.PortSet {
	exposedPorts := make(network.PortSet)

	for _, port := range namedPorts {
		np, ok := network.PortFrom(uint16(port.Port), network.IPProtocol(port.Protocol))
		if !ok {
			continue
		}
		exposedPorts[np] = struct{}{}
	}

	return exposedPorts
}

func (d *docker) portBindings(exposedPorts network.PortSet, ports NamedPorts) network.PortMap {
	portBindings := make(network.PortMap)

	// for the container to be accessible from another container, it cannot
	// listen on 127.0.0.1 as it will be accessed by gateway address (e.g
	// 172.17.0.1), so its port should be exposed everywhere
	hostAddr := d.hostAddr()
	if isInDocker() {
		hostAddr = "0.0.0.0"
	}

	hostIP := netip.MustParseAddr(hostAddr)

	for port := range exposedPorts {
		binding := network.PortBinding{
			HostIP: hostIP,
		}

		if pName, err := ports.Find(string(port.Proto()), int(port.Num())); err == nil {
			namedPort := ports.Get(pName)
			if namedPort.HostPort > 0 {
				binding.HostPort = strconv.Itoa(namedPort.HostPort)
			}
		}

		portBindings[port] = []network.PortBinding{binding}
	}

	return portBindings
}

func (d *docker) createContainer(
	ctx context.Context,
	image string,
	ports NamedPorts,
	cfg *Options,
) (*client.ContainerCreateResult, error) {
	exposedPorts := d.exposedPorts(ports)
	containerConfig := &container.Config{
		ExposedPorts: exposedPorts,
		Env:          cfg.Env,
		User:         cfg.User,
	}

	if len(cfg.Cmd) > 0 {
		containerConfig.Cmd = cfg.Cmd
	}

	if len(cfg.Entrypoint) > 0 {
		containerConfig.Entrypoint = cfg.Entrypoint
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
		AutoRemove:   !cfg.Debug,
		Privileged:   cfg.Privileged,
		Mounts:       mounts,
		ExtraHosts:   cfg.ExtraHosts,
	}

	createOpts := client.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
		Name:       cfg.ContainerName,
		Image:      image,
	}

	resp, err := d.client.ContainerCreate(ctx, createOpts)
	if err == nil {
		return &resp, nil
	}

	matches := duplicateContainerRegexp.FindStringSubmatch(err.Error())
	if len(matches) == 2 {
		d.log.Infow("duplicate container found, stopping", "container", matches[1])

		_, err = d.client.ContainerRemove(ctx, matches[1], client.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return nil, fmt.Errorf("can't remove existing container: %w", err)
		}

		resp, err = d.client.ContainerCreate(ctx, createOpts)
	}

	return &resp, err
}

func (d *docker) findReusableContainer(
	ctx context.Context,
	image string,
	ports NamedPorts,
	cfg *Options,
) (*Container, bool, error) {
	if cfg.ContainerName == "" {
		return nil, false, fmt.Errorf("container name is required when container reuse is enabled")
	}

	result, err := d.client.ContainerList(ctx, client.ContainerListOptions{
		Filters: make(client.Filters).
			Add("name", cfg.ContainerName).
			Add("ancestor", image).
			Add("status", "running"),
	})
	if err != nil || len(result.Items) < 1 {
		return nil, false, err
	}

	container, err := d.waitForContainerNetwork(ctx, result.Items[0].ID, ports)
	if err != nil {
		return nil, false, err
	}

	return container, true, nil
}

func (d *docker) boundNamedPorts(inspectResp container.InspectResponse, namedPorts NamedPorts) (NamedPorts, error) {
	boundNamedPorts := make(NamedPorts)

	for containerPort, bindings := range inspectResp.NetworkSettings.Ports {
		if len(bindings) == 0 {
			continue
		}

		hostPortNum, err := strconv.Atoi(bindings[0].HostPort)
		if err != nil {
			return nil, fmt.Errorf("invalid host port value '%s': %w", bindings[0].HostPort, err)
		}

		proto := string(containerPort.Proto())
		portNum := int(containerPort.Num())

		portName, err := namedPorts.Find(proto, portNum)
		if err != nil {
			return nil, fmt.Errorf("can't find port %s/%d: %w", proto, portNum, err)
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

	logsOptions := client.ContainerLogsOptions{
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

	stopTimeout := defaultStopTimeoutSec

	_, err := d.client.ContainerStop(ctx, id, client.ContainerStopOptions{
		Timeout: &stopTimeout,
	})
	if err != nil && !cerrdefs.IsNotFound(err) {
		return fmt.Errorf("can't stop container %s: %w", id, err)
	}

	return nil
}

func (d *docker) stopClient() error {
	return d.client.Close()
}

func (d *docker) removeContainer(ctx context.Context, id string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	_, err := d.client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{Force: true})
	if err != nil && !cerrdefs.IsNotFound(err) && !isDeletionAlreadyInProgessError(err, id) {
		return fmt.Errorf("can't remove container %s: %w", id, err)
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

func isDeletionAlreadyInProgessError(err error, id string) bool {
	return err.Error() == fmt.Sprintf("Error response from daemon: removal of container %s is already in progress", id)
}
