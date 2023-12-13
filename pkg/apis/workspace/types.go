package workspace

const (
	DefaultBlock         = "default"
	ProjectSelectorField = "projectSelector"

	BackendLocal            = "local"
	BackendMysql            = "mysql"
	BackendOss              = "oss"
	BackendS3               = "s3"
	EnvBackendMysqlPassword = "KUSION_BACKEND_MYSQL_PASSWORD"
	EnvAwsAccessKeyID       = "AWS_ACCESS_KEY_ID"
	EnvAwsSecretAccessKey   = "AWS_SECRET_ACCESS_KEY"
	EnvAwsDefaultRegion     = "AWS_DEFAULT_REGION"
	EnvAwsRegion            = "AWS_REGION"
	EnvOssAccessKeyID       = "OSS_ACCESS_KEY_ID"
	EnvOssAccessKeySecret   = "OSS_ACCESS_KEY_SECRET"
	DefaultMysqlPort        = 3306
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
type ModuleConfigs map[string]*ModuleConfig

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
type ModuleConfig struct {
	// Default is default block of the module config.
	Default GenericConfig `yaml:"default" json:"default"`

	// ModulePatcherConfigs are the patcher blocks of the module config.
	ModulePatcherConfigs `yaml:",inline,omitempty" json:",inline,omitempty"`
}

// ModulePatcherConfigs is a group of ModulePatcherConfig.
type ModulePatcherConfigs map[string]*ModulePatcherConfig

// ModulePatcherConfig is a patcher block of the module config.
type ModulePatcherConfig struct {
	// GenericConfig contains the module configs.
	GenericConfig `yaml:",inline" json:",inline"`

	// ProjectSelector contains the selected projects.
	ProjectSelector []string `yaml:"projectSelector" json:"projectSelector"`
}

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
type TerraformConfig map[string]*ProviderConfig

// ProviderConfig contains the full configurations of a specified provider. It is the combination
// of the specified provider's config in blocks "terraform/required_providers" and "providers" in
// terraform hcl file, where the former is described by fields Source and Version, and the latter
// is described by GenericConfig cause different provider has different config.
type ProviderConfig struct {
	// Source of the provider.
	Source string `yaml:"source" json:"source"`

	// Version of the provider.
	Version string `yaml:"version" json:"version"`

	// GenericConfig is used to describe the config of a specified terraform provider.
	GenericConfig `yaml:",inline,omitempty" json:",inline,omitempty"`
}

// BackendConfigs contains config of the backend, which is used to store state, etc. Only one kind
// backend can be configured.
type BackendConfigs struct {
	// Local is the backend using local file system.
	Local *LocalFileConfig `yaml:"local,omitempty" json:"local,omitempty"`

	// Mysql is the backend using mysql database.
	Mysql *MysqlConfig `yaml:"mysql,omitempty" json:"mysql,omitempty"`

	// Oss is the backend using OSS.
	Oss *OssConfig `yaml:"oss,omitempty" json:"oss,omitempty"`

	// S3 is the backend using S3.
	S3 *S3Config `yaml:"s3,omitempty" json:"s3,omitempty"`
}

// LocalFileConfig contains the config of using local file system as backend. Now there is no configuration
// item for local file.
type LocalFileConfig struct{}

// MysqlConfig contains the config of using mysql database as backend.
type MysqlConfig struct {
	// DBName is the database name.
	DBName string `yaml:"dbName" json:"dbName"`

	// User of the database.
	User string `yaml:"user" json:"user"`

	// Password of the database.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Host of the database.
	Host string `yaml:"host" json:"host"`

	// Port of the database. If not set, then it will be set to DefaultMysqlPort.
	Port *int `yaml:"port,omitempty" json:"port,omitempty"`
}

// OssConfig contains the config of using OSS as backend.
type OssConfig struct {
	GenericObjectStorageConfig `yaml:",inline" json:",inline"` // OSS asks for non-empty endpoint
}

// S3Config contains the config of using S3 as backend.
type S3Config struct {
	GenericObjectStorageConfig `yaml:",inline" json:",inline"`

	// Region of S3.
	Region string `yaml:"region,omitempty" json:"region,omitempty"`
}

// GenericObjectStorageConfig contains generic configs which can be reused by OssConfig and S3Config.
type GenericObjectStorageConfig struct {
	// Endpoint of the object storage service.
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// AccessKeyID of the object storage service.
	AccessKeyID string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty"`

	// AccessKeySecret of the object storage service.
	AccessKeySecret string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty"`

	// Bucket of the object storage service.
	Bucket string `yaml:"bucket" json:"bucket"`
}

// GenericConfig is a generic model to describe config which shields the difference among multiple concrete
// models. GenericConfig is designed for extensibility, used for module, terraform runtime config, etc.
type GenericConfig map[string]any
