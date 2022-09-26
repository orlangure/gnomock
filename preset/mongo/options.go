package mongo

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithData sets up initial container state according to the directory
// structure at the given path:
//
//   - path/first/one
//   - path/first/two
//   - path/second/three
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
// ignored.
func WithData(path string) Option {
	return func(p *P) {
		p.DataPath = path
	}
}

// WithUser creates a root user with the provided name and password. This user
// should be used as a part of mongodb connection string. If you choose not to
// use your own user and password, the databases will be unprotected, and you
// won't need to specify any name and password in your connection string.
func WithUser(user, pass string) Option {
	return func(p *P) {
		p.User = user
		p.Password = pass
	}
}

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}
