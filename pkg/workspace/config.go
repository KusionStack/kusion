package workspace

import (
	"kusionstack.io/kusion/pkg/models"
)

// Config is used to describe the config of a workspace.
type Config struct {
	// Modules are the config of a set of modules.
	Modules map[string]models.ModuleConfigs `yaml:"modules,omitempty" json:"modules,omitempty"`

	// Runtimes are the config of a set of runtimes.
	Runtimes map[string]models.RuntimeConfig `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`

	// Backends are the config of a set of backends.
	Backends map[string]models.BackendConfig `yaml:"backends,omitempty" json:"backends,omitempty"`
}

/*
// unstructuredConfig is the Config without specified structure, used as the intermedia between
// YAML file and structured Config.
type unstructuredConfig struct {
	Modules  map[string]any `yaml:"modules,omitempty" json:"modules,omitempty"`
	Runtimes map[string]any `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`
	Backends map[string]any `yaml:"backends,omitempty" json:"backends,omitempty"`
}
*/

/*
todo: the following functions are to get provided.
func ParseConfig(data []byte) (*Config, error) {}
func ParseConfigFromYamlFile(path string) (*Config, error) {}
func (c *Config) Validate() error {}
func (c *Config) GetModule(projectName, moduleName string) (models.Module, error) {}
func (c *Config) GetRuntimeConfig(runtimeName string) (models.RuntimeConfig, error) {}
func (c *Config) GetRuntimeConfigs() (map[string]models.RuntimeConfig, error) {}
func (c *Config) GetBackendConfig() (string, models.BackendConfig, error) {}
*/
