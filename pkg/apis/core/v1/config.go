package v1

const (
	BackendTypeLocal = "local"
	BackendTypeMysql = "mysql"
	BackendTypeOss   = "oss"
	BackendTypeS3    = "s3"
)

// Config contains configurations for kusion cli, which stores in ${KUSION_HOME}/config.yaml.
type Config struct {
	// Backends contains the configurations for multiple backends.
	Backends *BackendConfigs `yaml:"backends,omitempty"`
}

// BackendConfigs contains the configuration of multiple backends and the current backend.
type BackendConfigs struct {
	// Current is the name of the current used backend.
	Current string `yaml:"current,omitempty"`

	// Backends contains the types and configs of multiple backends, whose key is the backend name.
	Backends map[string]*BackendConfig `yaml:"backends,omitempty,inline"`
}

// BackendConfig contains the type and configs of a backend, which is used to store State and Workspace.
type BackendConfig struct {
	// Type is the backend type, supports BackendTypeLocal, BackendTypeMysql, BackendTypeOss, BackendTypeS3.
	Type string `yaml:"type,omitempty"`

	// Configs contains config items of the backend, whose keys differ from different backend types.
	Configs map[string]string `yaml:"configs,omitempty"`
}
