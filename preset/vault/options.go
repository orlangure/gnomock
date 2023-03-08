package vault

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}

// WithAuthToken sets authentication (root) token to be used to connect to this
// container.
func WithAuthToken(token string) Option {
	return func(p *P) {
		p.AuthToken = token
	}
}

// WithAuth enables new vault authorizations endpoints.
func WithAuth(auth []Auth) Option {
	return func(p *P) {
		p.Auth = auth
	}
}

// WithPolicies configures vault with the provided policies.
func WithPolicies(policies []Policy) Option {
	return func(p *P) {
		p.Policies = policies
	}
}

// WithAdditionalToken creates an additional access token with the provided policies and stores it
// in the provied file path.
func WithAdditionalToken(tc TokenCreate) Option {
	return func(p *P) {
		p.TokenCreate = &tc
	}
}
