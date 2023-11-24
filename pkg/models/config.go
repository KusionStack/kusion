package models

// ModuleConfig is an interface to describe the config of a module, which is organized in the
// ModuleConfigs to be provided in the workspace.Config.
type ModuleConfig struct {
	Module

	// ProjectSelector is used to define the projects that consume the ModuleConfig.
	ProjectSelector []string `yaml:"projectSelector,omitempty" json:"projectSelector,omitempty"`
}

// NewModuleConfig is used to new a ModuleConfig by a specified Module.
func NewModuleConfig(module Module) *ModuleConfig {
	return &ModuleConfig{Module: module}
}

// ModuleConfigs is a group of ModuleConfig.
type ModuleConfigs map[string]*ModuleConfig

// BackendConfig is an interface to describe the config of a backend.
type BackendConfig interface {
	// BackendName returns the name to identify the backend uniquely.
	BackendName() string
}

// SecureBackendConfig is an interface to describe the verified config of a backend.
type SecureBackendConfig interface {
	BackendConfig

	// Validate validates the BackendConfig is correct or not.
	Validate() error
}

// RuntimeConfig is an interface to describe the config of a runtime.
type RuntimeConfig interface {
	// RuntimeName returns the name to identify the runtime uniquely.
	RuntimeName() string
}

// SecureRuntimeConfig is an interface to describe the verified config of a runtime.
type SecureRuntimeConfig interface {
	RuntimeConfig

	// Validate validates the RuntimeConfig is correct or not.
	Validate() error
}
