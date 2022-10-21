package gnomock

// Preset is a type that includes ready to use Gnomock configuration. Such
// configuration includes image and ports as well as options specific to this
// implementation. For example, well known services like Redis or Postgres may
// have Gnomock implementations, providing healthcheck functions and basic
// initialization options.
type Preset interface {
	// Image returns a canonical docker image used to setup this Preset.
	Image() string

	// Ports returns a group of ports exposed by this Preset, where every port
	// is given a unique name. For example, if a container exposes API endpoint
	// on port 8080, and web interface on port 80, there should be two named
	// ports: "web" and "api".
	Ports() NamedPorts

	// Options returns a list of Option functions that allow to setup this
	// Preset implementation.
	Options() []Option
}

// ResetFunc is a function that can be called to reset container state between
// test provided to the WithContainerReset option.
//
// Some presets might implement a generic ResetFunc that should be fine for
// most use cases. For custom usage, it can be implemented depending on a
// use-case.
type ResetFunc func(*Container) error
