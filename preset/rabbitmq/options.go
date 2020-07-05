package rabbitmq

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*P)

// WithUser creates a new superuser with the provided credentials in the
// container
func WithUser(user, password string) Option {
	return func(p *P) {
		p.User = user
		p.Password = password
	}
}

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}
