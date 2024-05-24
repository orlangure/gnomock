package kafka

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}

// WithTopics makes sure that the provided topics are available when Kafka is
// up and running.
// Both topics from WithTopics and WithTopicConfigs will be added to Kafka.
func WithTopics(topics ...string) Option {
	return func(o *P) {
		o.Topics = append(o.Topics, topics...)
	}
}

// WithTopicConfigs makes sure that the provided topics with the given configs are available when Kafka is
// up and running. Unlike WithTopics, this allows to also set partition count.
// Both topics from WithTopics and WithTopicConfigs will be added to Kafka.
func WithTopicConfigs(topics ...TopicConfig) Option {
	return func(o *P) {
		o.TopicConfigs = append(o.TopicConfigs, topics...)
	}
}

// WithMessages makes sure that these messages can be consumed during the test
// once the container is ready.
func WithMessages(messages ...Message) Option {
	return func(o *P) {
		o.Messages = append(o.Messages, messages...)
	}
}

// WithMessagesFile allows to load messages to be sent into Kafka from one or
// multiple files.
func WithMessagesFile(files string) Option {
	return func(o *P) {
		o.MessagesFiles = append(o.MessagesFiles, files)
	}
}

// WithSchemaRegistry makes the container wait for the schema registry port to
// become available. Note that it takes longer to setup schema registry than
// the broker itself. Gnomock will not wait for the registry by default, but it
// may become available eventually.
func WithSchemaRegistry() Option {
	return func(o *P) {
		o.UseSchemaRegistry = true
	}
}
