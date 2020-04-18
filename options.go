package mysql

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*preset)

// WithUser creates a new superuser with the provided credentials in the
// container. If not used, the default credentials are gnomock:gnomick
func WithUser(user, password string) Option {
	return func(p *preset) {
		p.user = user
		p.password = password
	}
}

// WithDatabase creates a database with the provided name in the container. If
// not provided, "mydb" is used by default.  WithQueries, if provided, runs
// against the new database
func WithDatabase(db string) Option {
	return func(p *preset) {
		p.db = db
	}
}

// WithQueries executes the provided queries against the database created with
// WithDatabase, or against default "mydb" database
func WithQueries(queries ...string) Option {
	return func(p *preset) {
		p.queries = append(p.queries, queries...)
	}
}
