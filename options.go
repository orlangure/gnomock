package mongo

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*options)

// WithData sets up initial container state according to the directory
// structure at the given path:
//
//	- path:
// 	  - first
// 	    - one
// 	    - two
// 	  - second
// 	    - three
//
// For such directory structure, two databases are created: "first" and
// "second". Under "first" database there are two collections, "one" and "two",
// and under "second" database - one collection "three".
//
// Files "one", "two" and "three" are text files with JSON documents to be
// inserted into the database. One line should include one document.
//
// Top level files under "path" are ignored, only directories are used.
// Similarly, directories located anywhere besides top-level "path", are also
// ignored
func WithData(path string) Option {
	return func(o *options) {
		o.dataPath = path
	}
}

// WithUser creates a root user with the provided name and password. This user
// should be used as a part of mongodb connection string. If you choose not to
// use your own user and password, the databases will be unprotected, and you
// won't need to specify any name and password in your connection string
func WithUser(user, pass string) Option {
	return func(o *options) {
		o.user = user
		o.password = pass
	}
}

type options struct {
	dataPath string
	user     string
	password string
}

func buildConfig(opts ...Option) *options {
	config := &options{}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
