// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"time"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// Project is a definition of Kusion project resource.
//
// A project is composed of one or more applications and is linked to a Git repository(monorepo or polyrepo),
// which contains the project's desired intent.
type Project struct {
	// Name is a required fully qualified name.
	Name string `yaml:"name" json:"name"`

	// Description is an optional informational description.
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`

	// Labels is the list of labels that are assigned to this project.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`

	// Path is a directory path within the Git repository.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Stacks that belong to this project.
	Stacks []*Stack `yaml:"stacks,omitempty" json:"stacks,omitempty"`

	// Extensions allow you to customize how resources are generated of this project.
	Extensions []*Extension `yaml:"extensions,omitempty" json:"extensions,omitempty"`
}

// Stack is a definition of Kusion stack resource.
//
// Stack provides a mechanism to isolate multiple deployments of same application, it's the target workspace
// where application will be deployed to, the smallest operation unit that can be operated independently.
type Stack struct {
	// Name is a required fully qualified name.
	Name string `yaml:"name" json:"name"`

	// Backend is the place to store the workspace config and versioned releases of a stack.
	Backend string `yaml:"backend" json:"backend"`

	// Workspace is the target environment to deploy a stack.
	Workspace string `yaml:"workspace" json:"workspace"`

	// Description is an optional informational description.
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`

	// Labels is the list of labels that are assigned to this stack.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`

	// Path is a directory path within the Git repository.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Extensions allow you to customize how resources are generated of this project.
	Extensions []*Extension `yaml:"extensions,omitempty" json:"extensions,omitempty"`
}

const (
	DefaultBlock         = "default"
	ProjectSelectorField = "projectSelector"
)

// Workspace is a logical concept representing a target that stacks will be deployed to.
//
// Workspace is managed by platform engineers, which contains a set of configurations
// that application developers do not want or should not concern, and is reused by multiple
// stacks belonging to different projects.
type Workspace struct {
	// Name identifies a Workspace uniquely.
	Name string `yaml:"-" json:"-"`

	// Modules are the configs of a set of modules.
	Modules ModuleConfigs `yaml:"modules,omitempty" json:"modules,omitempty"`

	// SecretStore represents a secure external location for storing secrets.
	SecretStore *SecretStore `yaml:"secretStore,omitempty" json:"secretStore,omitempty"`

	// Context contains workspace-level configurations, such as topologies, server endpoints, metadata, etc.
	Context GenericConfig `yaml:"context,omitempty" json:"context,omitempty"`
}

const (
	BuiltinModulePrefix = ""
	ProbePrefix         = "service.container.probe."
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

const (
	TypeJob     = "job.Job"
	TypeService = "service.Service"
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
//	           "mysql" : d.MySQL {
//	               type: "cloud"
//	               version: "8.0"
//	           }
//	           # Built-in module, key represents the module source
//	           "prometheus" : m.Prometheus {
//	               path: "/metrics"
//	           }
//	           # Customized module, key represents the module source
//	           "customize": customizedModule {
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
	// The key in this map represents the module name
	Accessories map[string]Accessory `json:"accessories,omitempty" yaml:"accessories,omitempty"`
	// Labels and Annotations can be used to attach arbitrary metadata as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// Patcher primarily contains patches for fields associated with Workloads, and additionally offers the capability to patch other resources.
type Patcher struct {
	// Environments represent the environment variables patched to all containers in the workload.
	Environments []v1.EnvVar `json:"environments,omitempty" yaml:"environments,omitempty"`
	// Labels represent the labels patched to the workload.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// PodLabels represent the labels patched to the pods.
	PodLabels map[string]string `json:"podLabels,omitempty" yaml:"podLabels,omitempty"`
	// Annotations represent the annotations patched to the workload.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// PodAnnotations represent the annotations patched to the pods.
	PodAnnotations map[string]string `json:"podAnnotations,omitempty" yaml:"podAnnotations,omitempty"`
	// JSONPatchers represents patchers that can be patched to an arbitrary resource.
	// The key of this map represents the ResourceId of the resource to be patched.
	JSONPatchers map[string]JSONPatcher `json:"jsonPatcher,omitempty" yaml:"jsonPatcher,omitempty"`
}

type PatchType string

const (
	MergePatch PatchType = "MergePatch"
	JSONPatch  PatchType = "JSONPatch"
)

// JSONPatcher represents the patcher that can be patched to an arbitrary resource.
// The patch algorithm follows the RFC6902 JSON patch and RFC7396 JSON merge patches.
type JSONPatcher struct {
	// PatchType
	Type PatchType `json:"type" yaml:"type"`
	// Payload is the patch content.
	//
	// JSONPatch Example:
	// original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	// payload := []byte(`[
	//		{"op": "replace", "path": "/name", "value": "Jane"},
	//		{"op": "remove", "path": "/height"}
	//	]`)
	// result: {"age":24,"name":"Jane"}
	//
	// MergePatch Example:
	// original := []byte(`{"name": "Tina", "age": 28, "height": 3.75}`)
	// payload := []byte(`{"height":null,"name":"Jane"}`)
	// result: {"age":28,"name":"Jane"}
	Payload []byte `json:"payload" yaml:"payload"`
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
	BackendGenericOssEndpoint = "endpoint"
	BackendGenericOssAK       = "accessKeyID"
	BackendGenericOssSK       = "accessKeySecret"
	BackendGenericOssBucket   = "bucket"
	BackendGenericOssPrefix   = "prefix"
	BackendS3Region           = "region"

	BackendTypeLocal = "local"
	BackendTypeOss   = "oss"
	BackendTypeS3    = "s3"

	EnvOssAccessKeyID     = "OSS_ACCESS_KEY_ID"
	EnvOssAccessKeySecret = "OSS_ACCESS_KEY_SECRET"
	EnvAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	EnvAwsDefaultRegion   = "AWS_DEFAULT_REGION"
	EnvAwsRegion          = "AWS_REGION"
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
	// Type is the backend type, supports BackendTypeLocal, BackendTypeOss, BackendTypeS3.
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

// ModuleConfigs is a set of multiple ModuleConfig, whose key is the module name.
type ModuleConfigs map[string]*ModuleConfig

// GenericConfig is a generic model to describe config which shields the difference among multiple concrete
// models. GenericConfig is designed for extensibility, used for module, terraform runtime config, etc.
type GenericConfig map[string]any

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
// Take the ModuleConfig of "mysql" for an example, which is shown as below:
//
//	config := ModuleConfig{
//		"path":    "ghcr.io/kusionstack/mysql"
//		"version": "0.1.0"
//		"configs": {
//			"default": {
//				"type":         "aws",
//				"version":      "5.7",
//				"instanceType": "db.t3.micro",
//			},
//			"smallClass": {
//				"instanceType":    "db.t3.small",
//				"projectSelector": []string{"foo", "bar"},
//			},
//		},
//	}
type ModuleConfig struct {
	// Path is the path of the module. It can be a local path or a remote URL
	Path string `yaml:"path" json:"path"`
	// Version is the version of the module.
	Version string `yaml:"version" json:"version"`
	// Configs contains all levels of module configs
	Configs Configs `yaml:"configs" json:"configs"`
}

type Configs struct {
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

type ExtensionKind string

const (
	KubernetesMetadata  ExtensionKind = "kubernetesMetadata"
	KubernetesNamespace ExtensionKind = "kubernetesNamespace"
)

// Extension allows you to customize how resources are generated or customized as part of deployment.
type Extension struct {
	// Kind is a string value representing the extension.
	Kind ExtensionKind `yaml:"kind" json:"kind"`

	// The KubeNamespaceExtension
	KubeNamespace KubeNamespaceExtension `yaml:"kubernetesNamespace,omitempty" json:"kubernetesNamespace,omitempty"`

	// The KubeMetadataExtension
	KubeMetadata KubeMetadataExtension `yaml:"kubernetesMetadata,omitempty" json:"kubernetesMetadata,omitempty"`
}

// KubeNamespaceExtension allows you to override kubernetes namespace.
type KubeNamespaceExtension struct {
	// The custom namespace name
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

// KubeMetadataExtension allows you to append labels&annotations to kubernetes resources.
type KubeMetadataExtension struct {
	// Labels to add to kubernetes resources.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`

	// Annotations to add to kubernetes resources.
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// ExternalSecretRef contains information that points to the secret store data location.
type ExternalSecretRef struct {
	// Specifies the name of the secret in Provider to read, mandatory.
	Name string `yaml:"name" json:"name"`

	// Specifies the version of the secret to return, if supported.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`

	// Used to select a specific property of the secret data (if a map), if supported.
	Property string `yaml:"property,omitempty" json:"property,omitempty"`
}

// SecretStore contains configuration to describe target secret store.
type SecretStore struct {
	Provider *ProviderSpec `yaml:"provider" json:"provider"`
}

// ProviderSpec contains provider-specific configuration.
type ProviderSpec struct {
	// Alicloud configures a store to retrieve secrets from Alicloud Secrets Manager.
	Alicloud *AlicloudProvider `yaml:"alicloud,omitempty" json:"alicloud,omitempty"`

	// AWS configures a store to retrieve secrets from AWS Secrets Manager.
	AWS *AWSProvider `yaml:"aws,omitempty" json:"aws,omitempty"`

	// Vault configures a store to retrieve secrets from HashiCorp Vault.
	Vault *VaultProvider `yaml:"vault,omitempty" json:"vault,omitempty"`

	// Azure configures a store to retrieve secrets from Azure KeyVault.
	Azure *AzureKVProvider `yaml:"azure,omitempty" json:"azure,omitempty"`

	// Fake configures a store with static key/value pairs
	Fake *FakeProvider `yaml:"fake,omitempty" json:"fake,omitempty"`

	// Onprem configures a store in on-premises environments
	OnPremises *OnPremisesProvider `yaml:"onpremises,omitempty" json:"onpremises,omitempty"`
}

// AlicloudProvider configures a store to retrieve secrets from Alicloud Secrets Manager.
type AlicloudProvider struct {
	// Alicloud Region to be used to interact with Alicloud Secrets Manager.
	// Examples are cn-beijing, cn-shanghai, etc.
	Region string `yaml:"region" json:"region"`
}

// AWSProvider configures a store to retrieve secrets from AWS Secrets Manager.
type AWSProvider struct {
	// AWS Region to be used to interact with AWS Secrets Manager.
	// Examples are us-east-1, us-west-2, etc.
	Region string `yaml:"region" json:"region"`

	// The profile to be used to interact with AWS Secrets Manager.
	// If not set, the default profile created with `aws configure` will be used.
	Profile string `yaml:"profile,omitempty" json:"profile,omitempty"`
}

type VaultKVStoreVersion string

const (
	VaultKVStoreV1 VaultKVStoreVersion = "v1"
	VaultKVStoreV2 VaultKVStoreVersion = "v2"
)

// VaultProvider configures a store to retrieve secrets from HashiCorp Vault.
type VaultProvider struct {
	// Server is the target Vault server address to connect, e.g: "https://vault.example.com:8200".
	Server string `yaml:"server" json:"server"`

	// Path is the mount path of the Vault KV backend endpoint, e.g: "secret".
	Path *string `yaml:"path,omitempty" json:"path,omitempty"`

	// Version is the Vault KV secret engine version. Version can be either "v1" or
	// "v2", defaults to "v2".
	Version VaultKVStoreVersion `yaml:"version" json:"version"`
}

// AzureEnvironmentType specifies the Azure cloud environment endpoints to use for connecting and authenticating with Azure.
type AzureEnvironmentType string

const (
	AzureEnvironmentPublicCloud       AzureEnvironmentType = "PublicCloud"
	AzureEnvironmentUSGovernmentCloud AzureEnvironmentType = "USGovernmentCloud"
	AzureEnvironmentChinaCloud        AzureEnvironmentType = "ChinaCloud"
	AzureEnvironmentGermanCloud       AzureEnvironmentType = "GermanCloud"
)

// AzureKVProvider configures a store to retrieve secrets from Azure KeyVault
type AzureKVProvider struct {
	// Vault Url from which the secrets to be fetched from.
	VaultURL *string `yaml:"vaultUrl" json:"vaultUrl"`

	// TenantID configures the Azure Tenant to send requests to.
	TenantID *string `yaml:"tenantId" json:"tenantId"`

	// EnvironmentType specifies the Azure cloud environment endpoints to use for connecting and authenticating with Azure.
	// By-default it points to the public cloud AAD endpoint, and the following endpoints are available:
	// PublicCloud, USGovernmentCloud, ChinaCloud, GermanCloud
	// Ref: https://github.com/Azure/go-autorest/blob/main/autorest/azure/environments.go#L152
	EnvironmentType AzureEnvironmentType `yaml:"environmentType,omitempty" json:"environmentType,omitempty"`
}

// FakeProvider configures a fake provider that returns static values.
type FakeProvider struct {
	Data []FakeProviderData `json:"data"`
}

type FakeProviderData struct {
	Key      string            `json:"key"`
	Value    string            `json:"value,omitempty"`
	ValueMap map[string]string `json:"valueMap,omitempty"`
	Version  string            `json:"version,omitempty"`
}

// OnPremisesProvider configures a secret provider in on-premises environments
type OnPremisesProvider struct {
	// platform name of the provider
	Name string `json:"name"`
	// attributes of the provider
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Type string

const (
	Kubernetes Type = "Kubernetes"
	Terraform  Type = "Terraform"
)

const (
	// ResourceExtensionGVK is the key for resource extension, which is used to
	// store the GVK of the resource.
	ResourceExtensionGVK = "GVK"
	// ResourceExtensionKubeConfig is the key for resource extension, which is used
	// to indicate the path of kubeConfig for Kubernetes type resource.
	ResourceExtensionKubeConfig = "kubeConfig"
)

type Resources []Resource

// Resource is the representation of a resource in the state.
type Resource struct {
	// ID is the unique key of this resource.
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	// providerNamespace:providerName:resourceType:resourceName for Terraform resources
	ID string `yaml:"id" json:"id"`

	// Type represents all Context we supported like Kubernetes and Terraform
	Type Type `yaml:"type" json:"type"`

	// Attributes represents all specified attributes of this resource
	Attributes map[string]interface{} `yaml:"attributes" json:"attributes"`

	// DependsOn contains all resources this resource depends on
	DependsOn []string `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`

	// Extensions specifies arbitrary metadata of this resource
	Extensions map[string]interface{} `yaml:"extensions,omitempty" json:"extensions,omitempty"`
}

// Spec describes the desired state how the infrastructure should look like: which workload to run,
// the load-balancer setup, the location of the database schema, and so on. Based on that information,
// the Kusion engine takes care of updating the production state to match the Intent.
type Spec struct {
	// Resources is the list of Resource this Spec contains.
	Resources Resources `yaml:"resources" json:"resources"`
	// SecretSore represents a external secret store location for storing secrets.
	SecretStore *SecretStore `yaml:"secretStore" json:"secretStore"`
}

// State is a record of an operation's result. It is a mapping between resources in KCL and the actual
// infra resource and often used as a datasource for 3-way merge/diff in operations like Apply or Preview.
type State struct {
	// Resources records all resources in this operation.
	Resources Resources `yaml:"resources" json:"resources"`
}

// ReleasePhase is the Phase of a Release.
type ReleasePhase string

const (
	// ReleasePhaseGenerating indicates the stage of generating Spec.
	ReleasePhaseGenerating ReleasePhase = "generating"

	// ReleasePhasePreviewing indicated the stage of previewing.
	ReleasePhasePreviewing ReleasePhase = "previewing"

	// ReleasePhaseApplying indicates the stage of applying.
	ReleasePhaseApplying ReleasePhase = "applying"

	// ReleasePhaseDestroying indicates the stage of destroying.
	ReleasePhaseDestroying ReleasePhase = "destroying"

	// ReleasePhaseSucceeded is a final phase, indicates the Release is successful.
	ReleasePhaseSucceeded ReleasePhase = "succeeded"

	// ReleasePhaseFailed is a final phase, indicates the Release is failed.
	ReleasePhaseFailed ReleasePhase = "failed"
)

// Release describes the generation, preview and deployment of a specified Stack. When the operation
// Apply or Destroy is executed, a Release will be created.
type Release struct {
	// Project name of the Release.
	Project string `yaml:"project" json:"project"`

	// Workspace name of the Release.
	Workspace string `yaml:"workspace" json:"workspace"`

	// Revision of the Release, auto-increasing from one under per Project and Workspace. The group of
	// Project, Workspace and Revision can identify a Release uniquely.
	Revision uint64 `yaml:"revision" json:"revision"`

	// Stack name of the Release.
	Stack string `yaml:"stack" json:"stack"`

	// Spec of the Release, which can be provided when creating release or generated during Release.
	Spec *Spec `yaml:"spec,omitempty" json:"spec,omitempty"`

	// State of the Release, which will be generated and updated during Release. When a Release is created,
	// the State will be filled with the latest State, which indicates the current infra resources.
	State *State `yaml:"state" json:"state"`

	// Phase is the current phase of the Release.
	Phase ReleasePhase `yaml:"phase" json:"phase"`

	// CreateTime is the time that the Release is created.
	CreateTime time.Time `yaml:"createTime" json:"createTime"`

	// ModifiedTime is the time that the Release is modified.
	ModifiedTime time.Time `yaml:"modifiedTime" json:"modifiedTime"`
}
