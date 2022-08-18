package states

import "github.com/zclconf/go-cty/cty"

// Backend represent a medium that Kusion will operate on.
type Backend interface {
	// ConfigSchema returns a set of attributes that is needed to config this backend
	ConfigSchema() cty.Type

	// Configure will config this backend with provided configuration
	Configure(obj cty.Value) error

	// StateStorage return a StateStorage to manage State
	StateStorage() StateStorage
}
