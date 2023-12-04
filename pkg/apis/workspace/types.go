package workspace

const (
	DefaultBlock         = "default"
	ProjectSelectorField = "projectSelector"
)

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
	Runtimes *RuntimeConfigs `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`

	// Backends are the configs of a set of backends.
	Backends *BackendConfigs `yaml:"backends,omitempty" json:"backends,omitempty"`
}

// ModuleConfigs is a set of multiple ModuleConfig, whose key is the module name.
type ModuleConfigs map[string]ModuleConfig

// ModuleConfig is the config of a module, which contains a default and several patcher blocks.
//
// The default block's key is "default", and value is the module inputs. The patcher blocks' keys
// are the patcher names, which are just block identifiers without specific meaning, but must
// not be "default". Besides module inputs, patcher block's value also contains a field named
// "projectSelector", whose value is a slice containing the project names that use the patcher
// configs. A project can only be assigned in a patcher's "projectSelector" field, the assignment
// in multiple patchers is not allowed. For a project, if not specified in the patcher block's
// "projectSelector" field, the default config will be used.
//
// Take the ModuleConfig of "database" for an example, which is shown as below:
//
//	 config := ModuleConfig {
//		"default": {
//			"type":         "aws",
//			"version":      "5.7",
//			"instanceType": "db.t3.micro",
//		},
//		"smallClass": {
//		 	"instanceType":    "db.t3.small",
//		 	"projectSelector": []string{"foo", "bar"},
//		},
//	}
type ModuleConfig map[string]GenericConfig

// RuntimeConfigs contains a set of runtime config.
type RuntimeConfigs struct {
	// Kubernetes contains the config to access a kubernetes cluster.
	Kubernetes *KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`

	// Terraform contains the config of multiple terraform providers.
	Terraform TerraformConfig `yaml:"terraform,omitempty" json:"terraform,omitempty"`
}

// KubernetesConfig contains config to access a kubernetes cluster.
type KubernetesConfig struct {
	// KubeConfig is the path of the kubeconfig file.
	KubeConfig string `yaml:"kubeConfig" json:"kubeConfig"`
}

// TerraformConfig contains the config of multiple terraform provider config, whose key is
// the provider name.
type TerraformConfig map[string]GenericConfig

// BackendConfigs contains config of the backend, which is used to store state, etc. Only one kind
// backend can be configured.
// todo: add more backends declared in pkg/engine/backend
type BackendConfigs struct {
	// Local is backend to use local file system.
	Local *LocalFileConfig `yaml:"local,omitempty" json:"local,omitempty"`
}

// LocalFileConfig contains the config of using local file system as backend.
type LocalFileConfig struct {
	// Path is place to store state, etc.
	Path string `yaml:"path" json:"path"`
}

// GenericConfig is a generic model to describe config which shields the difference among multiple concrete
// models. GenericConfig is designed for extensibility, used for module, terraform runtime config, etc.
type GenericConfig map[string]any
