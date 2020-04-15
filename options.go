package mssql

const defaultPassword = "Gn0m!ck~"
const defaultDatabase = "mydb"
const defaultPort = 1433

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*options)

// WithAdminPassword sets administrator password that can be used to connect
// (default: Gn0m!ck~)
func WithAdminPassword(password string) Option {
	return func(o *options) {
		o.password = password
	}
}

// WithDatabase creates a database with the provided name in the container. If
// not provided, "mydb" is used by default.  WithQueries, if provided, runs
// against the new database
func WithDatabase(db string) Option {
	return func(o *options) {
		o.db = db
	}
}

// WithQueries executes the provided queries against the database created with
// WithDatabase, or against default "mydb" database
func WithQueries(queries ...string) Option {
	return func(o *options) {
		o.queries = append(o.queries, queries...)
	}
}

// WithLicense sets EULA acceptance state. To accept the license, use true. See
// https://hub.docker.com/_/microsoft-mssql-server?tab=description for more
// information
func WithLicense(accept bool) Option {
	return func(o *options) {
		o.license = accept
	}
}

type options struct {
	db       string
	queries  []string
	password string
	license  bool
}

func buildConfig(opts ...Option) *options {
	config := &options{
		db:       defaultDatabase,
		password: defaultPassword,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
