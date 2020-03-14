package splunk

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*options)

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

type options struct {
	values        []Event
	acceptLicense bool
	adminPassword string
}

func buildConfig(opts ...Option) *options {
	config := &options{}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
