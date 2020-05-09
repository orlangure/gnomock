// Package gnomock contains a framework to set up temporary docker containers
// for integration and end-to-end testing of other applications. It handles
// pulling images, starting containers, waiting for them to become available,
// setting up their initial state and cleaning up in the end.
//
// It can be used either directly, or via already existing implementations of
// various connectors built by the community.
package gnomock

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/sync/errgroup"
)

const defaultTag = "latest"

// StartCustom creates a new container using provided image and binds random
// ports on the host to the provided ports inside the container. Image may
// include tag, which is set to "latest" by default. Optional configuration is
// available through Option functions. The returned container must be stopped
// when no longer needed using its Stop() method
func StartCustom(image string, ports NamedPorts, opts ...Option) (c *Container, err error) {
	config := buildConfig(opts...)
	image = buildImage(image, config.Tag)

	startCtx, cancel := context.WithTimeout(config.ctx, config.StartTimeout)
	defer cancel()

	cli, err := dockerConnect()
	if err != nil {
		return nil, fmt.Errorf("can't create docker client: %w", err)
	}

	err = cli.pullImage(startCtx, image)
	if err != nil {
		return nil, fmt.Errorf("can't pull image: %w", err)
	}

	c, err = cli.startContainer(startCtx, image, ports, config)
	if err != nil {
		return nil, fmt.Errorf("can't start container: %w", err)
	}

	logReader, err := cli.readLogs(context.Background(), c.ID)
	if err != nil {
		return nil, fmt.Errorf("can't setup log forwarding: %w", err)
	}

	g := &errgroup.Group{}

	g.Go(copy(config.logWriter, logReader))

	c.onStop = closeLogReader(logReader, g)

	waitCtx, cancelWait := context.WithTimeout(config.ctx, config.WaitTimeout)
	defer cancelWait()

	err = wait(waitCtx, c, config)
	if err != nil {
		return c, fmt.Errorf("can't connect to container: %w", err)
	}

	err = config.init(c)
	if err != nil {
		return c, fmt.Errorf("can't init container: %w", err)
	}

	return c, nil
}

func copy(dst io.Writer, src io.Reader) func() error {
	return func() error {
		_, err := stdcopy.StdCopy(dst, dst, src)
		if err != nil && err != io.EOF {
			return err
		}

		return nil
	}
}

func closeLogReader(logReader io.ReadCloser, g *errgroup.Group) func() error {
	return func() error {
		err := logReader.Close()
		if err != nil {
			return err
		}

		err = g.Wait()
		if err != nil {
			return err
		}

		return nil
	}
}

// Start creates a container using the provided Preset. The Preset provides its
// own Options to configure Gnomock container. Usually this is enough, but it
// is still possible to extend/override Preset options with new values. For
// example, wait timeout defined in the Preset, if at all, might be not enough
// for this particular usage, so it can't be changed during this call.
//
// All provided Options are applied. First, Preset options are applied. Then,
// custom Options. If both Preset and custom Options change the same
// configuration, custom Options are used
func Start(p Preset, opts ...Option) (*Container, error) {
	presetOpts := p.Options()

	mergedOpts := make([]Option, 0, len(opts)+len(presetOpts))
	mergedOpts = append(mergedOpts, presetOpts...)
	mergedOpts = append(mergedOpts, opts...)

	return StartCustom(p.Image(), p.Ports(), mergedOpts...)
}

// Stop stops the provided container and lets docker remove them from the
// system. Stop returns an error if any one of the containers couldn't stop
func Stop(cs ...*Container) error {
	var g errgroup.Group

	for _, c := range cs {
		container := c

		g.Go(func() error {
			return stop(container)
		})
	}

	return g.Wait()
}

func stop(c *Container) error {
	if c == nil {
		return nil
	}

	cli, err := dockerConnect()
	if err != nil {
		return fmt.Errorf("can't create docker client: %w", err)
	}

	err = cli.stopContainer(context.Background(), c.ID)
	if err != nil {
		return fmt.Errorf("can't stop container: %w", err)
	}

	if c.onStop != nil {
		err = c.onStop()
		if err != nil {
			return fmt.Errorf("can't perform last cleanup: %w", err)
		}
	}

	return nil
}

func buildImage(image, tag string) string {
	parts := strings.Split(image, ":")

	noTagSet := len(parts) == 1
	if noTagSet {
		if tag == "" {
			tag = defaultTag
		}

		image = fmt.Sprintf("%s:%s", image, tag)
	} else if tag != "" {
		image = fmt.Sprintf("%s:%s", parts[0], tag)
	}

	return image
}

func wait(ctx context.Context, c *Container, config *Options) error {
	delay := time.NewTicker(config.healthcheckInterval)
	defer delay.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("canceled after error: %w", lastErr)
		case <-delay.C:
			err := config.healthcheck(c)
			if err == nil {
				return nil
			}

			lastErr = err
		}
	}
}
