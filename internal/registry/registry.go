// Package registry provides access to existing presets. Every preset is
// required to call `Register` in order to become discoverable in the registry.
package registry

import (
	"github.com/orlangure/gnomock"
)

type presetFactory func() gnomock.Preset

var registry = map[string]presetFactory{}

// Register makes the provided preset discoverable by the provided name.
func Register(name string, p presetFactory) {
	registry[name] = p
}

// Find returns a preset registered under the provided name, or nil if such
// name is not registered.
func Find(name string) gnomock.Preset {
	p, ok := registry[name]
	if !ok {
		return nil
	}

	return p()
}
