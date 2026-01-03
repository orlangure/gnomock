package jaeger

// Option is a configuration of Jaeger preset
type Option func(*P)

// WithVersion sets the version of Jaeger image to use
func WithVersion(version string) Option {
	return func(p *P) {
		p.Version = version
	}
}
