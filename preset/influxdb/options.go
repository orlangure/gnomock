package influxdb

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}

// WithUser sets super-user name and password for this container.
func WithUser(username, password string) Option {
	return func(p *P) {
		p.Username = username
		p.Password = password
	}
}

// WithOrg sets organization name for this container.
func WithOrg(org string) Option {
	return func(p *P) {
		p.Org = org
	}
}

// WithBucket sets the initial bucket name for this container.
func WithBucket(bucket string) Option {
	return func(p *P) {
		p.Bucket = bucket
	}
}

// WithAuthToken sets authentication token to be used to connect to this
// container.
func WithAuthToken(token string) Option {
	return func(p *P) {
		p.AuthToken = token
	}
}
