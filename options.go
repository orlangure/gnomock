package redis

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*preset)

// WithValues initializes Redis with the provided key/value pairs. These values
// never expire. See go-redis/redis package for information on supported value
// types
func WithValues(vs map[string]interface{}) Option {
	return func(p *preset) {
		p.initialValues = vs
	}
}
