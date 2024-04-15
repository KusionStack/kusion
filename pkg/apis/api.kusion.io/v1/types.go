package v1

import (
	"time"

	"kusionstack.io/kusion/pkg/version"
)

// Project is a definition of Kusion project resource.
//
// A project is composed of one or more applications and is linked to a Git repository(monorepo or polyrepo),
// which contains the project's desired intent.
type Project struct {
	// Name is a required fully qualified name.
	Name string `json:"name" yaml:"name"`
	// Description is an optional informational description.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels is the list of labels that are assigned to this project.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// Path is a directory path within the Git repository.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// The set of stacks that are known about this project.
	Stacks []*Stack `json:"stacks,omitempty" yaml:"stacks,omitempty"`
}

// Stack is a definition of Kusion stack resource.
//
// Stack provides a mechanism to isolate multiple deployments of same application, it's the target workspace
// where application will be deployed to, the smallest operation unit that can be operated independently.
type Stack struct {
	// Name is a required fully qualified name.
	Name string `json:"name" yaml:"name"`
	// Backend is the place to store the workspace config and versioned releases of a stack.
	Backend string `json:"backend" yaml:"backend"`
	// Workspace is the target environment to deploy a stack.
	Workspace string `json:"workspace" yaml:"workspace"`
	// Description is an optional informational description.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels is the list of labels that are assigned to this stack.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// Path is a directory path within the Git repository.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
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

	// Runtimes are the configs of a set of runtimes.
	Runtimes *RuntimeConfigs `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`

	// SecretStore represents a secure external location for storing secrets.
	SecretStore *SecretStoreSpec `yaml:"secretStore,omitempty" json:"secretStore,omitempty"`
}

// ModuleConfigs is a set of multiple ModuleConfig, whose key is the module name.
// The module name format is "source@version".
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

type VaultKVStoreVersion string

const (
	VaultKVStoreV1 VaultKVStoreVersion = "v1"
	VaultKVStoreV2 VaultKVStoreVersion = "v2"
)

// ExternalSecretRef contains information that points to the secret store data location.
type ExternalSecretRef struct {
	// Specifies the name of the secret in Provider to read, mandatory.
	Name string `yaml:"name" json:"name"`

	// Specifies the version of the secret to return, if supported.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`

	// Used to select a specific property of the secret data (if a map), if supported.
	Property string `yaml:"property,omitempty" json:"property,omitempty"`
}

// SecretStoreSpec contains configuration to describe target secret store.
type SecretStoreSpec struct {
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
	// ID is the unique key of this resource in the whole DeprecatedState.
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	// providerNamespace:providerName:resourceType:resourceName for Terraform resources
	ID string `json:"id" yaml:"id"`

	// Type represents all Runtimes we supported like Kubernetes and Terraform
	Type Type `json:"type" yaml:"type"`

	// Attributes represents all specified attributes of this resource
	Attributes map[string]interface{} `json:"attributes" yaml:"attributes"`

	// DependsOn contains all resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	// Extensions specifies arbitrary metadata of this resource
	Extensions map[string]interface{} `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}

// Spec describes the desired state how the infrastructure should look like: which workload to run,
// the load-balancer setup, the location of the database schema, and so on. Based on that information,
// the Kusion engine takes care of updating the production state to match the Intent.
type Spec struct {
	// Resources is the list of Resource this Spec contains.
	Resources Resources `yaml:"resources" json:"resources"`
}

// State is a record of an operation's result. It is a mapping between resources in KCL and the actual infra
// resource and often used as a datasource for 3-way merge/diff in operations like Apply or Preview.
type State struct {
	// Resources are the result returned from the engine runtime.
	Resources Resources `yaml:"resources" json:"resources"`
}

// ReleaseStatus is the status of a release
type ReleaseStatus string

const (
	// ReleaseStatusGenerating indicates the stage of generating Spec.
	ReleaseStatusGenerating = "generating"

	// ReleaseStatusPreviewing indicated the stage of previewing.
	ReleaseStatusPreviewing = "previewing"

	// ReleaseStatusDeploying indicates the stage of deploying.
	ReleaseStatusDeploying = "deploying"

	// ReleaseStatusSucceeded is a final state, indicates the Release is successful.
	ReleaseStatusSucceeded = "succeeded"

	// ReleaseStatusFailed is a final state, indicates the Release is failed.
	ReleaseStatusFailed = "failed"
)

// Release describes the generation, preview and deployment of a specified Stack. When a kusion
// apply is executed, a Release will be created.
type Release struct {
	// Name is used to identify a release uniquely, which is generated by Kusion automatically.
	Name string `yaml:"name" json:"name"`

	// Project name of the Release.
	Project string `yaml:"project" json:"project"`

	// Stack name of the Release.
	Stack string `yaml:"stack" json:"stack"`

	// Spec of the Release, which can be provided when creating release or generated during Release.
	Spec *Spec `yaml:"spec,omitempty" json:"spec,omitempty"`

	// State of the Release, which will be generated and updated during Release.
	State *State `yaml:"state,omitempty" json:"state,omitempty"`

	// Status is the current state of the Release.
	Status ReleaseStatus `yaml:"status" json:"status"`

	// CreateTime is the time that the Release is created.
	CreateTime time.Time `yaml:"createTime" json:"createTime"`

	// ModifiedTime is the time that the Release is modified.
	ModifiedTime time.Time `yaml:"modifiedTime,omitempty" json:"modifiedTime,omitempty"`
}

// Deprecated: DeprecatedState will not in use in v0.11.2
type DeprecatedState struct {
	// DeprecatedState ID
	ID int64 `json:"id" yaml:"id"`

	// Project name
	Project string `json:"project" yaml:"project"`

	// Stack name
	Stack string `json:"stack" yaml:"stack"`

	// Workspace name
	Workspace string `json:"workspace" yaml:"workspace"`

	// DeprecatedState version
	Version int `json:"version" yaml:"version"`

	// KusionVersion represents the Kusion's version when this DeprecatedState is created
	KusionVersion string `json:"kusionVersion" yaml:"kusionVersion"`

	// Serial is an auto-increase number that represents how many times this DeprecatedState is modified
	Serial uint64 `json:"serial" yaml:"serial"`

	// Operator represents the person who triggered this operation
	Operator string `json:"operator,omitempty" yaml:"operator,omitempty"`

	// Resources records all resources in this operation
	Resources Resources `json:"resources" yaml:"resources"`

	// CreateTime is the time DeprecatedState is created
	CreateTime time.Time `json:"createTime" yaml:"createTime"`

	// ModifiedTime is the time DeprecatedState is modified each time
	ModifiedTime time.Time `json:"modifiedTime,omitempty" yaml:"modifiedTime,omitempty"`
}

func NewState() *DeprecatedState {
	s := &DeprecatedState{
		KusionVersion: version.ReleaseVersion(),
		Version:       1,
		Resources:     []Resource{},
	}
	return s
}
