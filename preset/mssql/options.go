package mssql

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*P)

// WithAdminPassword sets administrator password that can be used to connect
// (default: Gn0m!ck~)
func WithAdminPassword(password string) Option {
	return func(o *P) {
		o.Password = password
	}
}

// WithDatabase creates a database with the provided name in the container. If
// not provided, "mydb" is used by default.  WithQueries, if provided, runs
// against the new database
func WithDatabase(db string) Option {
	return func(o *P) {
		o.DB = db
	}
}

// WithQueries executes the provided queries against the database created with
// WithDatabase, or against default "mydb" database
func WithQueries(queries ...string) Option {
	return func(o *P) {
		o.Queries = append(o.Queries, queries...)
	}
}

// WithLicense sets EULA acceptance state. To accept the license, use true. See
// https://hub.docker.com/_/microsoft-mssql-server?tab=description for more
// information
func WithLicense(accept bool) Option {
	return func(o *P) {
		o.License = accept
	}
}

// WithQueriesFile sets a file name to read initial queries from. Queries from
// this file are executed before any other queries provided in WithQueries
func WithQueriesFile(file string) Option {
	return func(p *P) {
		p.QueriesFiles = append(p.QueriesFiles, file)
	}
}

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}
