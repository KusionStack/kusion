package workspace

// BackendConfigs contains config of the backend, which is used to store state, etc. Only one kind
// backend can be configured.
// todo: add more backends declared in pkg/engine/states
type BackendConfigs struct {
	// Local is backend to use local file system.
	Local LocalBackend `yaml:"local,omitempty" json:"local,omitempty"`
}

// LocalBackend contains the config of using local file system as backend.
type LocalBackend struct {
	// Path is place to store state, etc.
	Path string `yaml:"path" json:"path"`
}
