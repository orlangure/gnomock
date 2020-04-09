package mysql

const defaultUser = "gnomock"
const defaultPassword = "gnomick"
const defaultDatabase = "mydb"
const defaultPort = 3306

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*options)

// WithUser creates a new superuser with the provided credentials in the
// container. If not used, the default credentials are gnomock:gnomick
func WithUser(user, password string) Option {
	return func(o *options) {
		o.user = user
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

type options struct {
	db       string
	queries  []string
	user     string
	password string
}

func buildConfig(opts ...Option) *options {
	config := &options{
		db:       defaultDatabase,
		user:     defaultUser,
		password: defaultPassword,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
