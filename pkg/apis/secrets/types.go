package secrets

// ExternalSecretRef contains information that points to the secret store data location.
type ExternalSecretRef struct {
	// Specifies the path of the secret to read.
	Path string `yaml:"path" json:"path"`

	// Used to select a specific property of the secret data (if a map), if supported.
	Property string `yaml:"property,omitempty" json:"property,omitempty"`

	// Specifies the version of the secret to return, if supported.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// SecretStoreSpec contains configuration to describe target secret store.
type SecretStoreSpec struct {
	Provider *ProviderSpec `yaml:"provider" json:"provider"`
}

// ProviderSpec contains provider-specific configuration.
type ProviderSpec struct {
	// AWS configures a store to retrieve secrets from AWS Secrets Manager.
	AWS *AWSProvider `yaml:"aws,omitempty" json:"aws,omitempty"`
	// Vault configures a store to retrieve secrets from HashiCorp Vault.
	Vault *VaultProvider `yaml:"vault,omitempty" json:"vault,omitempty"`
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
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}
