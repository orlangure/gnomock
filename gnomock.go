package gnomock

import (
	"context"
	"fmt"
	"time"
)

const defaultVersion = "latest"
const defaultStartTimeout = time.Second * 60
const defaultStopTimeout = time.Second
const defaultWaitTimeout = time.Second * 10
const defaultHealthcheckInterval = time.Millisecond * 50

func Start(opts Options) (c *Container, err error) {
	opts, err = prepareOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	startCtx, cancel := context.WithTimeout(opts.Ctx, opts.StartTimeout)
	defer cancel()

	cli, err := dockerConnect()
	if err != nil {
		return nil, fmt.Errorf("can't create docker client: %w", err)
	}

	image := fmt.Sprintf("%s:%s", opts.Image, opts.Version)

	err = cli.pullImage(startCtx, image)
	if err != nil {
		return nil, fmt.Errorf("can't pull image: %w", err)
	}

	c, err = cli.startContainer(startCtx, image, opts)
	if err != nil {
		return nil, fmt.Errorf("can't start container: %w", err)
	}

	waitCtx, cancelWait := context.WithTimeout(opts.Ctx, opts.WaitTimeout)
	defer cancelWait()

	err = wait(waitCtx, c, opts)
	if err != nil {
		return c, fmt.Errorf("can't connect to container: %w", err)
	}

	err = opts.Init(c)
	if err != nil {
		return c, fmt.Errorf("can't init container: %w", err)
	}

	return c, nil
}

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

func prepareOptions(opts Options) (Options, error) {
	if opts.Image == "" {
		return opts, ErrImageNotSet
	}

	if opts.Version == "" {
		opts.Version = defaultVersion
	}

	if opts.Port == 0 {
		return opts, ErrPortNotSet
	}

	if opts.HealthcheckInterval == 0 {
		opts.HealthcheckInterval = defaultHealthcheckInterval
	}

	if opts.Healthcheck == nil {
		opts.Healthcheck = nopHealthcheck
	}

	if opts.Init == nil {
		opts.Init = nopInit
	}

	if opts.Ctx == nil {
		opts.Ctx = context.Background()
	}

	if opts.StartTimeout == 0 {
		opts.StartTimeout = defaultStartTimeout
	}

	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = defaultWaitTimeout
	}

	return opts, nil
}

func wait(ctx context.Context, c *Container, opts Options) error {
	delay := time.NewTicker(opts.HealthcheckInterval)
	defer delay.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("canceled after error: %w", lastErr)
		case <-delay.C:
			err := opts.Healthcheck(c.Host, c.Port)
			if err == nil {
				return nil
			}

			lastErr = err
		}
	}
}
