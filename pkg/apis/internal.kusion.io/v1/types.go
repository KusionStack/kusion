package v1

import (
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

const (
	BuiltinModulePrefix = "kam."
	ProbePrefix         = "v1.workload.container.probe."
	TypeHTTP            = BuiltinModulePrefix + ProbePrefix + "Http"
	TypeExec            = BuiltinModulePrefix + ProbePrefix + "Exec"
	TypeTCP             = BuiltinModulePrefix + ProbePrefix + "Tcp"
)

// Container describes how the App's tasks are expected to be run.
type Container struct {
	// Image to run for this container
	Image string `yaml:"image" json:"image"`
	// Entrypoint array.
	// The image's ENTRYPOINT is used if this is not provided.
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
	// Arguments to the entrypoint.
	// The image's CMD is used if this is not provided.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Collection of environment variables to set in the container.
	// The value of environment variable may be static text or a value from a secret.
	Env yaml.MapSlice `yaml:"env,omitempty" json:"env,omitempty"`
	// The current working directory of the running process defined in entrypoint.
	WorkingDir string `yaml:"workingDir,omitempty" json:"workingDir,omitempty"`
	// Resource requirements for this container.
	Resources map[string]string `yaml:"resources,omitempty" json:"resources,omitempty"`
	// Files configures one or more files to be created in the container.
	Files map[string]FileSpec `yaml:"files,omitempty" json:"files,omitempty"`
	// Dirs configures one or more volumes to be mounted to the specified folder.
	Dirs map[string]string `yaml:"dirs,omitempty" json:"dirs,omitempty"`
	// Periodic probe of container liveness.
	LivenessProbe *Probe `yaml:"livenessProbe,omitempty" json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.
	ReadinessProbe *Probe `yaml:"readinessProbe,omitempty" json:"readinessProbe,omitempty"`
	// StartupProbe indicates that the Pod has successfully initialized.
	StartupProbe *Probe `yaml:"startupProbe,omitempty" json:"startupProbe,omitempty"`
	// Actions that the management system should take in response to container lifecycle events.
	Lifecycle *Lifecycle `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
}

// FileSpec defines the target file in a Container
type FileSpec struct {
	// The content of target file in plain text.
	Content string `yaml:"content,omitempty" json:"content,omitempty"`
	// Source for the file content, might be a reference to a secret value.
	ContentFrom string `yaml:"contentFrom,omitempty" json:"contentFrom,omitempty"`
	// Mode bits used to set permissions on this file.
	Mode string `yaml:"mode" json:"mode"`
}

// TypeWrapper is a thin wrapper to make YAML decoder happy.
type TypeWrapper struct {
	// Type of action to be taken.
	Type string `yaml:"_type" json:"_type"`
}

// Probe describes a health check to be performed against a container to determine whether it is
// alive or ready to receive traffic.
type Probe struct {
	// The action taken to determine the health of a container.
	ProbeHandler *ProbeHandler `yaml:"probeHandler" json:"probeHandler"`
	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" json:"initialDelaySeconds,omitempty"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `yaml:"timeoutSeconds,omitempty" json:"timeoutSeconds,omitempty"`
	// How often (in seconds) to perform the probe.
	PeriodSeconds int32 `yaml:"periodSeconds,omitempty" json:"periodSeconds,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	SuccessThreshold int32 `yaml:"successThreshold,omitempty" json:"successThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `yaml:"failureThreshold,omitempty" json:"failureThreshold,omitempty"`
}

// ProbeHandler defines a specific action that should be taken in a probe.
// One and only one of the fields must be specified.
type ProbeHandler struct {
	// Type of action to be taken.
	TypeWrapper `yaml:"_type" json:"_type"`
	// Exec specifies the action to take.
	// +optional
	*ExecAction `yaml:",inline" json:",inline"`
	// HTTPGet specifies the http request to perform.
	// +optional
	*HTTPGetAction `yaml:",inline" json:",inline"`
	// TCPSocket specifies an action involving a TCP port.
	// +optional
	*TCPSocketAction `yaml:",inline" json:",inline"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem.
	// Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// URL is the full qualified url location to send HTTP requests.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Custom headers to set in the request. HTTP allows repeated headers.
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

// TCPSocketAction describes an action based on opening a socket.
type TCPSocketAction struct {
	// URL is the full qualified url location to open a socket.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

// Lifecycle describes actions that the management system should take in response
// to container lifecycle events.
type Lifecycle struct {
	// PreStop is called immediately before a container is terminated due to an
	// API request or management event such as liveness/startup probe failure,
	// preemption, resource contention, etc.
	PreStop *LifecycleHandler `yaml:"preStop,omitempty" json:"preStop,omitempty"`
	// PostStart is called immediately after a container is created.
	PostStart *LifecycleHandler `yaml:"postStart,omitempty" json:"postStart,omitempty"`
}

// LifecycleHandler defines a specific action that should be taken in a lifecycle
// hook. One and only one of the fields, except TCPSocket must be specified.
type LifecycleHandler struct {
	// Type of action to be taken.
	TypeWrapper `yaml:"_type" json:"_type"`
	// Exec specifies the action to take.
	// +optional
	*ExecAction `yaml:",inline" json:",inline"`
	// HTTPGet specifies the http request to perform.
	// +optional
	*HTTPGetAction `yaml:",inline" json:",inline"`
}

type Protocol string

const (
	TCP Protocol = "TCP"
	UDP Protocol = "UDP"
)

// Port defines the exposed port of Service.
type Port struct {
	// Port is the exposed port of the Service.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
	// TargetPort is the backend .Container port.
	TargetPort int `yaml:"targetPort,omitempty" json:"targetPort,omitempty"`
	// Protocol is protocol used to expose the port, support ProtocolTCP and ProtocolUDP.
	Protocol Protocol `yaml:"protocol,omitempty" json:"protocol,omitempty"`
}

type Secret struct {
	Type      string            `yaml:"type" json:"type"`
	Params    map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
	Data      map[string]string `yaml:"data,omitempty" json:"data,omitempty"`
	Immutable bool              `yaml:"immutable,omitempty" json:"immutable,omitempty"`
}

const (
	FieldLabels      = "labels"
	FieldAnnotations = "annotations"
	FieldReplicas    = "replicas"
)

// Base defines set of attributes shared by different workload profile, e.g. Service and Job.
type Base struct {
	// The templates of containers to be run.
	Containers map[string]Container `yaml:"containers,omitempty" json:"containers,omitempty"`
	// The number of containers that should be run.
	Replicas *int32 `yaml:"replicas,omitempty" json:"replicas,omitempty"`
	// Secret
	Secrets map[string]Secret `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	// Dirs configures one or more volumes to be mounted to the specified folder.
	Dirs map[string]string `json:"dirs,omitempty" yaml:"dirs,omitempty"`
	// Labels and Annotations can be used to attach arbitrary metadata as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type ServiceType string

const (
	ModuleService                 = "service"
	ModuleServiceType             = "type"
	Deployment        ServiceType = "Deployment"
	Collaset          ServiceType = "CollaSet"
)

// Service is a kind of workload profile that describes how to run your application code.
// This is typically used for long-running web applications that should "never" go down, and handle short-lived latency-sensitive
// web requests, or events.
type Service struct {
	Base `yaml:",inline" json:",inline"`
	// Type represents the type of workload.Service, support Deployment and CollaSet.
	Type ServiceType `yaml:"type" json:"type"`
	// Ports describe the list of ports need getting exposed.
	Ports []Port `yaml:"ports,omitempty" json:"ports,omitempty"`
}

const ModuleJob = "job"

// Job is a kind of workload profile that describes how to run your application code. This is typically used for tasks that take from
// a few seconds to a few days to complete.
type Job struct {
	Base `yaml:",inline" json:",inline"`
	// The scheduling strategy in Cron format: https://en.wikipedia.org/wiki/Cron.
	Schedule string `yaml:"schedule,omitempty" json:"schedule,omitempty"`
}

type Type string

const (
	TypeJob     = "kam.v1.workload.Job"
	TypeService = "kam.v1.workload.Service"
)

type Header struct {
	Type string `yaml:"_type" json:"_type"`
}

type Workload struct {
	Header   `yaml:",inline" json:",inline"`
	*Service `yaml:",inline" json:",inline"`
	*Job     `yaml:",inline" json:",inline"`
}

type Accessory map[string]interface{}

// AppConfiguration is a developer-centric definition that describes how to run an App. The application model is built on a decade
// of experience from AntGroup in operating a large-scale internal developer platform and combines the best ideas and practices from the
// community.
//
// Note: AppConfiguration per se is not a Kusion Module
//
// Example:
// import models.schema.v1 as ac
// import models.schema.v1.workload as wl
// import models.schema.v1.workload.container as c
// import models.schema.v1.workload.container.probe as p
// import models.schema.v1.monitoring as m
// import models.schema.v1.database as d
//
//		helloWorld: ac.AppConfiguration {
//		   # Built-in module
//		   workload: wl.Service {
//		       containers: {
//		           "main": c.Container {
//		               image: "ghcr.io/kusion-stack/samples/helloworld:latest"
//		               # Configure a HTTP readiness probe
//		               readinessProbe: p.Probe {
//		                   probeHandler: p.Http {
//		                       url: "http://localhost:80"
//		                   }
//		               }
//		           }
//		       }
//		   }
//
//		   # extend accessories module base
//	       accessories: {
//	           # Built-in module, key represents the module source
//	           "kusionstack/mysql@v0.1" : d.MySQL {
//	               type: "cloud"
//	               version: "8.0"
//	           }
//	           # Built-in module, key represents the module source
//	           "kusionstack/prometheus@v0.1" : m.Prometheus {
//	               path: "/metrics"
//	           }
//	           # Customized module, key represents the module source
//	           "foo/customize": customizedModule {
//	               ...
//	           }
//	       }
//
//		   # extend pipeline module base
//		   pipeline: {
//		       # Step is a module
//		       "step" : Step {
//		           use: "exec"
//		           args: ["--test-all"]
//		       }
//		   }
//
//		   # Dependent app list
//		   dependency: {
//		       dependentApps: ["init-kusion"]
//		   }
//		}
type AppConfiguration struct {
	// Name of the target App.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Workload defines how to run your application code.
	Workload *Workload `json:"workload" yaml:"workload"`
	// Accessories defines a collection of accessories that will be attached to the workload.
	// The key in this map represents the module source. e.g. kusionstack/mysql@v0.1
	Accessories map[string]Accessory `json:"accessories,omitempty" yaml:"accessories,omitempty"`
	// Labels and Annotations can be used to attach arbitrary metadata as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// Patcher contains fields should be patched into the workload corresponding fields
type Patcher struct {
	// Environments represent the environment variables patched to all containers in the workload.
	Environments []v1.EnvVar `json:"environments" yaml:"environments"`
	// Labels represent the labels patched to both the workload and pod.
	Labels map[string]string `json:"labels" yaml:"labels"`
	// Annotations represent the annotations patched to both the workload and pod.
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

const ConfigBackends = "backends"

// Config contains configurations for kusion cli, which stores in ${KUSION_HOME}/config.yaml.
type Config struct {
	// Backends contains the configurations for multiple backends.
	Backends *BackendConfigs `yaml:"backends,omitempty" json:"backends,omitempty"`
}

const (
	DefaultBackendName = "default"

	BackendCurrent            = "current"
	BackendType               = "type"
	BackendConfigItems        = "configs"
	BackendLocalPath          = "path"
	BackendMysqlDBName        = "dbName"
	BackendMysqlUser          = "user"
	BackendMysqlPassword      = "password"
	BackendMysqlHost          = "host"
	BackendMysqlPort          = "port"
	BackendGenericOssEndpoint = "endpoint"
	BackendGenericOssAK       = "accessKeyID"
	BackendGenericOssSK       = "accessKeySecret"
	BackendGenericOssBucket   = "bucket"
	BackendGenericOssPrefix   = "prefix"
	BackendS3Region           = "region"

	BackendTypeLocal = "local"
	BackendTypeMysql = "mysql"
	BackendTypeOss   = "oss"
	BackendTypeS3    = "s3"

	EnvBackendMysqlPassword = "KUSION_BACKEND_MYSQL_PASSWORD"
	EnvOssAccessKeyID       = "OSS_ACCESS_KEY_ID"
	EnvOssAccessKeySecret   = "OSS_ACCESS_KEY_SECRET"
	EnvAwsAccessKeyID       = "AWS_ACCESS_KEY_ID"
	EnvAwsSecretAccessKey   = "AWS_SECRET_ACCESS_KEY"
	EnvAwsDefaultRegion     = "AWS_DEFAULT_REGION"
	EnvAwsRegion            = "AWS_REGION"

	DefaultMysqlPort = 3306
)

// BackendConfigs contains the configuration of multiple backends and the current backend.
type BackendConfigs struct {
	// Current is the name of the current used backend.
	Current string `yaml:"current,omitempty" json:"current,omitempty"`

	// Backends contains the types and configs of multiple backends, whose key is the backend name.
	Backends map[string]*BackendConfig `yaml:",omitempty,inline" json:",omitempty,inline"`
}

// BackendConfig contains the type and configs of a backend, which is used to store Spec, State and Workspace.
type BackendConfig struct {
	// Type is the backend type, supports BackendTypeLocal, BackendTypeMysql, BackendTypeOss, BackendTypeS3.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Configs contains config items of the backend, whose keys differ from different backend types.
	Configs map[string]any `yaml:"configs,omitempty" json:"configs,omitempty"`
}

// BackendLocalConfig contains the config of using local file system as backend, which can be converted
// from BackendConfig if Type is BackendTypeLocal.
type BackendLocalConfig struct {
	// Path of the directory to store the files.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// BackendMysqlConfig contains the config of using mysql database as backend, which can be converted
// from BackendConfig if Type is BackendMysqlConfig.
type BackendMysqlConfig struct {
	// DBName is the database name.
	DBName string `yaml:"dbName" json:"dbName"`

	// User of the database.
	User string `yaml:"user" json:"user"`

	// Password of the database.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Host of the database.
	Host string `yaml:"host" json:"host"`

	// Port of the database. If not set, then it will be set to DeprecatedDefaultMysqlPort.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
}

// BackendOssConfig contains the config of using OSS as backend, which can be converted from BackendConfig
// if Type is BackendOssConfig.
type BackendOssConfig struct {
	*GenericBackendObjectStorageConfig `yaml:",inline" json:",inline"` // OSS asks for non-empty endpoint
}

// BackendS3Config contains the config of using S3 as backend, which can be converted from BackendConfig
// if Type is BackendS3Config.
type BackendS3Config struct {
	*GenericBackendObjectStorageConfig `yaml:",inline" json:",inline"`

	// Region of S3.
	Region string `yaml:"region,omitempty" json:"region,omitempty"`
}

// GenericBackendObjectStorageConfig contains generic configs which can be reused by BackendOssConfig and
// BackendS3Config.
type GenericBackendObjectStorageConfig struct {
	// Endpoint of the object storage service.
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// AccessKeyID of the object storage service.
	AccessKeyID string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty"`

	// AccessKeySecret of the object storage service.
	AccessKeySecret string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty"`

	// Bucket of the object storage service.
	Bucket string `yaml:"bucket" json:"bucket"`

	// Prefix of the key to store the files.
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
}

// ToLocalBackend converts BackendConfig to structured BackendLocalConfig, works only when the Type
// is BackendTypeLocal, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToLocalBackend() *BackendLocalConfig {
	if b.Type != BackendTypeLocal {
		return nil
	}
	path, _ := b.Configs[BackendLocalPath].(string)
	return &BackendLocalConfig{
		Path: path,
	}
}

// ToMysqlBackend converts BackendConfig to structured BackendMysqlConfig, works only when the Type
// is BackendTypeMysql, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToMysqlBackend() *BackendMysqlConfig {
	if b.Type != BackendTypeMysql {
		return nil
	}
	dbName, _ := b.Configs[BackendMysqlDBName].(string)
	user, _ := b.Configs[BackendMysqlUser].(string)
	password, _ := b.Configs[BackendMysqlPassword].(string)
	host, _ := b.Configs[BackendMysqlHost].(string)
	port, _ := b.Configs[BackendMysqlPort].(int)
	return &BackendMysqlConfig{
		DBName:   dbName,
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
	}
}

// ToOssBackend converts BackendConfig to structured BackendOssConfig, works only when the Type is
// BackendTypeOss, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToOssBackend() *BackendOssConfig {
	if b.Type != BackendTypeOss {
		return nil
	}
	endpoint, _ := b.Configs[BackendGenericOssEndpoint].(string)
	accessKeyID, _ := b.Configs[BackendGenericOssAK].(string)
	accessKeySecret, _ := b.Configs[BackendGenericOssSK].(string)
	bucket, _ := b.Configs[BackendGenericOssBucket].(string)
	prefix, _ := b.Configs[BackendGenericOssPrefix].(string)
	return &BackendOssConfig{
		&GenericBackendObjectStorageConfig{
			Endpoint:        endpoint,
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			Bucket:          bucket,
			Prefix:          prefix,
		},
	}
}

// ToS3Backend converts BackendConfig to structured BackendS3Config, works only when the Type is
// BackendTypeS3, and the Configs are with correct type, or return nil.
func (b *BackendConfig) ToS3Backend() *BackendS3Config {
	if b.Type != BackendTypeS3 {
		return nil
	}
	endpoint, _ := b.Configs[BackendGenericOssEndpoint].(string)
	accessKeyID, _ := b.Configs[BackendGenericOssAK].(string)
	accessKeySecret, _ := b.Configs[BackendGenericOssSK].(string)
	bucket, _ := b.Configs[BackendGenericOssBucket].(string)
	prefix, _ := b.Configs[BackendGenericOssPrefix].(string)
	region, _ := b.Configs[BackendS3Region].(string)
	return &BackendS3Config{
		GenericBackendObjectStorageConfig: &GenericBackendObjectStorageConfig{
			Endpoint:        endpoint,
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			Bucket:          bucket,
			Prefix:          prefix,
		},
		Region: region,
	}
}
