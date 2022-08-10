package init

import (
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
)

// backends store all available backend
var backends map[string]func() states.Backend

// init backends map with all support backend
func init() {
	backends = map[string]func() states.Backend{
		"local": local.NewLocalBackend,
	}
}

// GetBackend return backend, or nil if not exists
func GetBackend(name string) func() states.Backend {
	return backends[name]
}
