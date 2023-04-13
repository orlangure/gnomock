package k3s

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}

// WithPort allows to use a custom port for k3s API access instead of the
// default one. If no custom port is provided, port 48443 is used instead.
//
// Please make sure that whichever port you choose to use (including the
// default) is available on the host system. Otherwise this container won't
// start.
func WithPort(port int) Option {
	return func(o *P) {
		o.Port = port
	}
}

// WithDynamicPort configures the preset to find and use a dynamic (free) port
// for k3s API access from the host. The kubeconfig is replaced with the
// container host port so local Kubernetes clients will still work.
//
// This is preferred to `WithPort`, which has been maintained for backward
// compatibility, as it does not require a specific host port to be available.
func WithDynamicPort() Option {
	return func(o *P) {
		o.UseDynamicPort = true
	}
}
