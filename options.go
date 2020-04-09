package postgres

import "fmt"

const defaultUser = "postgres"
const defaultPassword = "password"
const defaultDatabase = "postgres"
const defaultSSLMode = "disable"
const defaultPort = 5432

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*options)

// WithUser creates a new superuser with the provided credentials in the
// container
func WithUser(user, password string) Option {
	return func(o *options) {
		q := fmt.Sprintf(
			`create user %s with superuser password '%s'`,
			user, password,
		)

		o.queries = append(o.queries, q)
	}
}

// WithDatabase creates a database with the provided name in the container.
// WithQueries, if provided, runs against the new database
func WithDatabase(db string) Option {
	return func(o *options) {
		o.db = db
	}
}

// WithQueries executes the provided queries against the database created with
// WithDatabase, or against default postgres database
func WithQueries(queries ...string) Option {
	return func(o *options) {
		o.queries = append(o.queries, queries...)
	}
}

type options struct {
	db      string
	queries []string
}

func buildConfig(opts ...Option) *options {
	config := &options{
		db: defaultDatabase,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
