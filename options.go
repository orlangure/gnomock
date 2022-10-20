package gnomock

import (
	"context"
	"io"
	"time"
)

const (
	defaultTimeout             = time.Second * 300
	defaultHealthcheckInterval = time.Millisecond * 250
)

// Option is an optional Gnomock configuration. Functions implementing this
// signature may be combined to configure Gnomock containers for different use
// cases.
type Option func(*Options)

// WithContext sets the provided context to be used for setting up a Gnomock
// container. Canceling this context will cause Start() to abort.
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

// WithInit lets the provided InitFunc to be called when a Gnomock container is
// created, but before Start() returns. Use this function to run arbitrary code
// on the new container before using it. It can be useful to bring the
// container to a certain state (e.g create SQL schema).
func WithInit(f InitFunc) Option {
	return func(o *Options) {
		o.init = f
	}
}

// WithHealthCheck allows to define a rule to consider a Gnomock container
// ready to use. For example, it can attempt to connect to this container, and
// return an error on any failure, or nil on success. This function is called
// repeatedly until the timeout is reached, or until a nil error is returned.
func WithHealthCheck(f HealthcheckFunc) Option {
	return func(o *Options) {
		o.healthcheck = f
	}
}

// WithHealthCheckInterval defines an interval between two consecutive health
// check calls. This is a constant interval.
func WithHealthCheckInterval(t time.Duration) Option {
	return func(o *Options) {
		o.healthcheckInterval = t
	}
}

// WithTimeout sets the amount of time to wait for a created container to
// become ready to use. All startup steps must complete before they time out:
// start, wait until healthy, init.
func WithTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// WithEnv adds environment variable to the container. For example,
// `AWS_ACCESS_KEY_ID=FOOBARBAZ`.
func WithEnv(env string) Option {
	return func(o *Options) {
		o.Env = append(o.Env, env)
	}
}

// WithLogWriter sets the target where to write container logs. This can be
// useful for debugging.
func WithLogWriter(w io.Writer) Option {
	return func(o *Options) {
		o.logWriter = w
	}
}

// WithDebugMode allows Gnomock to output internal messages for debug purposes.
// Containers created in debug mode will not be automatically removed on
// failure to setup their initial state. Containers still might be removed if
// they are shut down from the inside. Use WithLogWriter to see what happens
// inside.
func WithDebugMode() Option {
	return func(o *Options) {
		o.Debug = true
	}
}

// WithContainerName allows to give a specific name to a new container. If a
// container with the same name already exists, it is killed.
func WithContainerName(name string) Option {
	return func(o *Options) {
		o.ContainerName = name
	}
}

// WithPrivileged starts a container in privileged mode (like `docker run
// --privileged`). This option should not be used unless you really need it.
// One use case for this option would be to run a Preset that has some kind of
// docker-in-docker functionality.
func WithPrivileged() Option {
	return func(o *Options) {
		o.Privileged = true
	}
}

// WithOptions allows to provide an existing set of Options instead of using
// optional configuration.
//
// This way has its own limitations. For example, context or initialization
// functions cannot be set in this way.
func WithOptions(options *Options) Option {
	return func(o *Options) {
		if options.Timeout > 0 {
			o.Timeout = options.Timeout
		}

		if options.CustomNamedPorts != nil {
			o.CustomNamedPorts = options.CustomNamedPorts
		}

		o.Env = append(o.Env, options.Env...)
		o.Debug = options.Debug
		o.ContainerName = options.ContainerName
	}
}

// WithCommand sets the command and its arguments to execute when container
// first runs. This command replaces the command defined in docker image.
func WithCommand(cmd string, args ...string) Option {
	return func(o *Options) {
		o.Cmd = append([]string{cmd}, args...)
	}
}

// WithHostMounts allows to bind host path (`src`) inside the container under
// `dst` path.
func WithHostMounts(src, dst string) Option {
	return func(o *Options) {
		if o.HostMounts == nil {
			o.HostMounts = make(map[string]string)
		}

		o.HostMounts[src] = dst
	}
}

// WithDisableAutoCleanup disables auto-removal of this container when the
// tests complete. Automatic cleanup is a safety net for tests that for some
// reason fail to run `gnomock.Stop()` in the end, for example due to an
// unexpected `os.Exit()` somewhere.
func WithDisableAutoCleanup() Option {
	return func(o *Options) {
		o.DisableAutoCleanup = true
	}
}

// WithUseLocalImagesFirst if possible to avoid hitting the Docker Hub pull rate limit.
func WithUseLocalImagesFirst() Option {
	return func(o *Options) {
		o.UseLocalImagesFirst = true
	}
}

// WithCustomNamedPorts allows to define custom ports for a container. This
// option should be used to override the ports defined by presets.
func WithCustomNamedPorts(namedPorts NamedPorts) Option {
	return func(o *Options) {
		o.CustomNamedPorts = namedPorts
	}
}

// WithRegistryAuth allows to access private docker images. The credentials
// should be passes as a Base64 encoded string, where the content is a JSON
// string with two fields: username and password.
//
// For Docker Hub, if 2FA authentication is enabled, an access token should be
// used instead of a password.
//
// For example: eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIn0K which stands
// for {"username":"foo","password":"bar"}.
func WithRegistryAuth(auth string) Option {
	return func(o *Options) {
		o.Auth = auth
	}
}

// WithContainerReuse disables Gnomock default behaviour of automatic container
// cleanup and also disables the automatic replacement at startup of an existing
// container with the same name and image. Effectively this makes Gnomock reuse
// a container from a previous Gnomock execution.
func WithContainerReuse() Option {
	return func(o *Options) {
		o.Reuse = true
	}
}

// HealthcheckFunc defines a function to be used to determine container health.
// It receives a host and a port, and returns an error if the container is not
// ready, or nil when the container can be used. One example of HealthcheckFunc
// would be an attempt to establish the same connection to the container that
// the application under test uses.
type HealthcheckFunc func(context.Context, *Container) error

func nopHealthcheck(context.Context, *Container) error {
	return nil
}

// InitFunc defines a function to be called on a ready to use container to set
// up its initial state before running the tests. For example, InitFunc can
// take care of creating a SQL table and inserting test data into it.
type InitFunc func(context.Context, *Container) error

func nopInit(context.Context, *Container) error {
	return nil
}

// Options includes Gnomock startup configuration. Functional options
// (WithSomething) should be used instead of directly initializing objects of
// this type whenever possible.
type Options struct {
	// Timeout is an amount of time to wait before considering Start operation
	// as failed.
	Timeout time.Duration `json:"timeout"`

	// Env is a list of environment variable to inject into the container. Each
	// entry is in format ENV_VAR_NAME=value
	Env []string `json:"env"`

	// Debug flag allows Gnomock to be verbose about steps it takes
	Debug bool `json:"debug"`

	// Privileged starts a container in privileged mode.
	Privileged bool `json:"privileged"`

	// ContainerName allows to use a specific name for a new container. In case
	// a container with the same name already exists, Gnomock kills it.
	ContainerName string `json:"container_name"`

	// Cmd is an optional command with its arguments to execute on container
	// startup. This command replaces the default one set on docker image
	// level.
	Cmd []string `json:"cmd"`

	// HostMounts allows to mount local paths into the container.
	HostMounts map[string]string `json:"host_mounts"`

	// DisableAutoCleanup prevents the container from being automatically
	// stopped and removed after the tests are complete. By default, Gnomock
	// will try to stop containers created by it right after the tests exit.
	DisableAutoCleanup bool `json:"disable_cleanup"`

	// WithUseLocalImagesFirst allows to use existing local images if possible
	// instead of always pulling the images.
	UseLocalImagesFirst bool `json:"use_local_images_first"`

	// CustomNamedPorts allows to override the ports set by the presets. This
	// option is useful for cases when the presets need to be created with
	// custom port definitions. This is an advanced feature and should be used
	// with care.
	//
	// Note that when using this option, you should provide custom named ports
	// with names matching the original ports returned by the used preset.
	//
	// When calling StartCustom directly from Go, it is possible to provide the
	// ports directly to the function.
	CustomNamedPorts NamedPorts `json:"custom_named_ports"`

	// Base64 encoded JSON string with docker access credentials. JSON string
	// should include two fields: username and password. For Docker Hub, if 2FA
	// authentication is enabled, an access token should be used instead of a
	// password.
	//
	// For example:
	//	eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIn0K
	// which stands for
	//	{"username":"foo","password":"bar"}
	Auth string `json:"auth"`

	// Reuse prevents the container from being automatically stopped and enables
	// its re-use in posterior executions.
	Reuse bool `json:"reuse"`

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
		Timeout:             defaultTimeout,
		logWriter:           io.Discard,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
