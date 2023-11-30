package workspace

// Workspace is a logical concept representing a target that stacks will be deployed to.
// Workspace is managed by platform engineers, which contains a set of configurations
// that application developers do not want or should not concern, and is reused by multiple
// stacks belonging to different projects.
type Workspace struct {
	// Name identifies a Workspace uniquely.
	Name string `yaml:"-" json:"-"`

	// Modules are the configs of a set of modules.
	Modules ModuleConfigs `yaml:"modules,omitempty" json:"modules,omitempty"`

	// Runtimes are the configs of a set of runtimes.
	Runtimes RuntimeConfigs `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`

	// Backends are the configs of a set of backends.
	Backends BackendConfigs `yaml:"backends,omitempty" json:"backends,omitempty"`
}
