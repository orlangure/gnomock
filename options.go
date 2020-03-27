package splunk

import "time"

const defaultInitTimeout = time.Second * 5

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*options)

// WithVersion sets splunk version (see
// https://hub.docker.com/r/splunk/splunk/tags) for available versions
func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

// WithValues initializes Splunk with the provided values as log entries
func WithValues(vs []Event) Option {
	return func(o *options) {
		o.values = vs
	}
}

// WithLicense lets the user choose to accept Splunk enterprise license
// (see more at https://hub.docker.com/_/splunk-enterprise). Failure to accept
// the license will prevent Splunk container from starting
func WithLicense(accept bool) Option {
	return func(o *options) {
		o.acceptLicense = accept
	}
}

// WithPassword sets admin password in Splunk container. Use this password to
// connect to the container when it is ready. Note that Splunk has password
// requirements. Failure to meet those will prevent the container from starting
// (see defaults at
// https://docs.splunk.com/Documentation/Splunk/latest/Security/Configurepasswordsinspecfile)
func WithPassword(pass string) Option {
	return func(o *options) {
		o.adminPassword = pass
	}
}

// WithInitTimeout sets the duration to wait before giving up on trying to
// initialize this Splunk container. This option is only useful when WithValues
// is used. Default value is 5 seconds
func WithInitTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.initTimeout = timeout
	}
}

type options struct {
	values        []Event
	acceptLicense bool
	adminPassword string
	initTimeout   time.Duration
	version       string
}

func buildConfig(opts ...Option) *options {
	config := &options{
		initTimeout: defaultInitTimeout,
		version:     "latest",
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
