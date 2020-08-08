package memcached

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*P)

// WithValues initializes Memcached with the provided key/value pairs. These values
// never expire. Only strings are supported.
func WithValues(vs map[string]string) Option {
	return func(p *P) {
		p.Values = vs
	}
}

// WithByteValues initializes Memcached with the provided key/value paris. These values
// never expire. Only byte slices are supported.
func WithByteValues(vs map[string][]byte) Option {
	return func(p *P) {
		p.ByteValues = vs
	}
}

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}
