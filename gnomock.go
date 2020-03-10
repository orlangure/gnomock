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
	"strings"
	"time"
)

const defaultTag = "latest"

// Start creates a new container using provided image and binds a random port
// on the host to the provided port inside the container. Image may include
// tag, which is set to "latest" by default. Optional configuration is
// available through Option functions. The returned container must be stopped
// when no longer needed using its Stop() method
func Start(image string, port int, opts ...Option) (c *Container, err error) {
	config, image := buildConfig(opts...), buildImage(image)

	startCtx, cancel := context.WithTimeout(config.ctx, config.startTimeout)
	defer cancel()

	cli, err := dockerConnect()
	if err != nil {
		return nil, fmt.Errorf("can't create docker client: %w", err)
	}

	err = cli.pullImage(startCtx, image)
	if err != nil {
		return nil, fmt.Errorf("can't pull image: %w", err)
	}

	c, err = cli.startContainer(startCtx, image, port)
	if err != nil {
		return nil, fmt.Errorf("can't start container: %w", err)
	}

	waitCtx, cancelWait := context.WithTimeout(config.ctx, config.waitTimeout)
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

// StartPreset creates a container using the provided Preset. For more
// information, see Start
func StartPreset(p Preset) (c *Container, err error) {
	return Start(p.Image(), p.Port(), p.Options()...)
}

// Stop stops this container and lets docker to remove it from the system
func Stop(c *Container) error {
	if c == nil {
		return nil
	}

	cli, err := dockerConnect()
	if err != nil {
		return fmt.Errorf("can't create docker client: %w", err)
	}

	ctx, timeout := context.Background(), defaultStopTimeout

	err = cli.stopContainer(ctx, c.ID, timeout)
	if err != nil {
		return fmt.Errorf("can't stop container: %w", err)
	}

	return nil
}

func buildImage(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) == 1 {
		image = fmt.Sprintf("%s:%s", image, defaultTag)
	}

	return image
}

func wait(ctx context.Context, c *Container, config *options) error {
	delay := time.NewTicker(config.healthcheckInterval)
	defer delay.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("canceled after error: %w", lastErr)
		case <-delay.C:
			err := config.healthcheck(c.Host, c.Port)
			if err == nil {
				return nil
			}

			lastErr = err
		}
	}
}
