package k3s

import (
	"strings"

	"golang.org/x/mod/semver"
)

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		// support original orlangure/k3s versions with rancher/k3s by
		// appending '-k3s1' to them.
		if version != "latest" {
			if !strings.Contains(version, "k3s") {
				version += "-k3s1"
			}

			if semver.Compare(version, "v1.17.0") > 0 {
				// --disable flag only exists in k3s versions above v1.17.0.
				o.K3sServerFlags = append(o.K3sServerFlags, "--disable=traefik")
			}
		}

		o.Version = version
	}
}

// Deprecated: WithPort allows to use a custom port for k3s API access instead
// of the default one. If no custom port is provided, port 48443 is used
// instead.
//
// Please make sure that whichever port you choose to use (including the
// default) is available on the host system. Otherwise this container won't
// start.
//
// This option and its affects has been kept as is for backward compatibility.
// We recommend using `WithDynamicPort()` instead as it does not require a
// static port to be available on the host.
func WithPort(port int) Option {
	return func(o *P) {
		o.Port = port
	}
}

// WithDynamicPort configures the preset to find and use a dynamic (free) port
// for k3s API access from the host. The kubeconfig is replaced with the
// container host port so local Kubernetes clients will still work.
//
// This is preferred to `WithPort` as it does not require a specific host port
// to be available.
func WithDynamicPort() Option {
	return func(o *P) {
		o.UseDynamicPort = true
	}
}
