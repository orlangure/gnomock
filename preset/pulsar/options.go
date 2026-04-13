package pulsar

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version. See
// https://hub.docker.com/r/apachepulsar/pulsar/tags for a list of available tags.
func WithVersion(version string) Option {
	return func(p *P) {
		p.Version = version
	}
}

// WithTopics makes sure that the provided topics are available when Pulsar is
// up and running.
func WithTopics(topics ...string) Option {
	return func(p *P) {
		p.Topics = append(p.Topics, topics...)
	}
}

// WithMessages makes sure that these messages can be consumed during the test
// once the container is ready.
func WithMessages(messages ...Message) Option {
	return func(p *P) {
		p.Messages = append(p.Messages, messages...)
	}
}
