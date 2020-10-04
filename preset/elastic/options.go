package elastic

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container.
type Option func(*P)

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}

// WithInputFile allows to load JSON data located in the provided file into
// Elasticsearch container once it is up. File name is used as ingestion index.
// Use multiple files to ingest JSON data into multiple indices.
//
// Data should be provided as a sequence of separate JSON objects, where each
// object is ingested as a single entity. There is no need to make an array of
// objects; a regular JSON object sequence without any separators will work.
func WithInputFile(file string) Option {
	return func(o *P) {
		o.Inputs = append(o.Inputs, file)
	}
}
