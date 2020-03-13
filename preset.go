package gnomock

// Preset is a type that includes ready to use Gnomock configuration. Such
// configuration includes image and port as well as options specific to this
// implementation. For example, well known services like Redis or Postgres may
// have Gnomock implementations, providing healthcheck functions and basic
// initialization options
type Preset interface {
	// Image returns a canonical docker image used to setup this Preset
	Image() string

	// Port returns a port number used by the container created with this
	// Preset
	Ports() NamedPorts

	// Options returns a list of Option functions that allow to setup this
	// Preset implementation
	Options() []Option
}
