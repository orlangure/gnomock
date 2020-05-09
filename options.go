package gnomock

import (
	"context"
	"io"
	"io/ioutil"
	"time"
)

const defaultStartTimeout = time.Second * 300
const defaultWaitTimeout = time.Second * 60
const defaultHealthcheckInterval = time.Millisecond * 250

// Option is an optional Gnomock configuration. Functions implementing this
// signature may be combined to configure Gnomock containers for different use
// cases
type Option func(*Options)

// WithContext sets the provided context to be used for setting up a Gnomock
// container. Canceling this context will cause Start() to abort
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

// WithInit lets the provided InitFunc to be called when a Gnomock container is
// created, but before Start() returns. Use this function to run arbitrary code
// on the new container before using it. It can be useful to bring the
// container to a certain state (e.g create SQL schema)
func WithInit(f InitFunc) Option {
	return func(o *Options) {
		o.init = f
	}
}

// WithHealthCheck allows to define a rule to consider a Gnomock container
// ready to use. For example, it can attempt to connect to this container, and
// return an error on any failure, or nil on success. This function is called
// repeatedly until the timeout is reached, or until a nil error is returned
func WithHealthCheck(f HealthcheckFunc) Option {
	return func(o *Options) {
		o.healthcheck = f
	}
}

// WithHealthCheckInterval defines an interval between two consecutive health
// check calls. This is a constant interval
func WithHealthCheckInterval(t time.Duration) Option {
	return func(o *Options) {
		o.healthcheckInterval = t
	}
}

// WithStartTimeout sets the amount of time to wait for a new Gnomock container
// to start. This includes pulling an image and creating a new container from
// it. To set the amount of time to wait before a created container healthy and
// ready to use, use WithWaitTimeout
func WithStartTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.StartTimeout = t
	}
}

// WithWaitTimeout sets the amount of time to wait for a created container to
// become ready to use. If health check function does not return nil error
// until this timeout is reached, Start() fails
func WithWaitTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.WaitTimeout = t
	}
}

// WithEnv adds environment variable to the container. For example,
// AWS_ACCESS_KEY_ID=FOOBARBAZ
func WithEnv(env string) Option {
	return func(o *Options) {
		o.Env = append(o.Env, env)
	}
}

// WithLogWriter sets the target where to write container logs. This can be
// useful for debugging
func WithLogWriter(w io.Writer) Option {
	return func(o *Options) {
		o.logWriter = w
	}
}

// WithTag overrides docker image tag provided by a preset
func WithTag(tag string) Option {
	return func(o *Options) {
		o.Tag = tag
	}
}

// WithOptions allows to provide an existing set of Options instead of using
// optional configuration.
//
// This way has its own limitations. For example, context or initialization
// functions cannot be set in this way
func WithOptions(options *Options) Option {
	return func(o *Options) {
		if options.StartTimeout > 0 {
			o.StartTimeout = options.StartTimeout
		}

		if options.WaitTimeout > 0 {
			o.WaitTimeout = options.WaitTimeout
		}

		if options.Tag != "" {
			o.Tag = options.Tag
		}

		o.Env = append(o.Env, options.Env...)
	}
}

// HealthcheckFunc defines a function to be used to determine container health.
// It receives a host and a port, and returns an error if the container is not
// ready, or nil when the container can be used. One example of HealthcheckFunc
// would be an attempt to establish the same connection to the container that
// the application under test uses
type HealthcheckFunc func(*Container) error

func nopHealthcheck(*Container) error {
	return nil
}

// InitFunc defines a function to be called on a ready to use container to set
// up its initial state before running the tests. For example, InitFunc can
// take care of creating a SQL table and inserting test data into it
type InitFunc func(*Container) error

func nopInit(*Container) error {
	return nil
}

// Options includes Gnomock startup configuration. Functional options
// (WithSomething) should be used instead of directly initializing objects of
// this type whenever possible
type Options struct {
	// StartTimeout is an amount of nanoseconds to wait for the container to be
	// created. This includes the time to pull docker image, create and start
	// the container. It does not include the time to wait for the container to
	// become healthy
	StartTimeout time.Duration `json:"start_timeout"`

	// WaitTimeout is an amount of nanoseconds to wait for a created container
	// to become healthy and ready to use
	WaitTimeout time.Duration `json:"wait_timeout"`

	// Env is a list of environment variable to inject into the container. Each
	// entry is in format ENV_VAR_NAME=value
	Env []string `json:"env"`

	// Tag sets docker image version to be used in this container. By default,
	// latest tag is used
	Tag string `json:"tag"`

	ctx                 context.Context
	init                InitFunc
	healthcheck         HealthcheckFunc
	healthcheckInterval time.Duration
	logWriter           io.Writer
}

func buildConfig(opts ...Option) *Options {
	config := &Options{
		ctx:                 context.Background(),
		init:                nopInit,
		healthcheck:         nopHealthcheck,
		healthcheckInterval: defaultHealthcheckInterval,
		StartTimeout:        defaultStartTimeout,
		WaitTimeout:         defaultWaitTimeout,
		logWriter:           ioutil.Discard,
		Tag:                 "",
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
