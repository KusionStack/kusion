package workspace

// RuntimeConfigs contains a set of runtime config.
type RuntimeConfigs struct {
	// Kubernetes contains the config to access a kubernetes cluster.
	Kubernetes KubernetesRuntime `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`

	// Terraform contains the config of multiple terraform providers.
	Terraform TerraformRuntime `yaml:"terraform,omitempty" json:"terraform,omitempty"`
}

// KubernetesRuntime contains config to access a kubernetes cluster.
type KubernetesRuntime struct {
	// KubeConfig is the path of the kubeconfig file.
	KubeConfig string `yaml:"kubeConfig" json:"kubeConfig"`
}

// TerraformRuntime contains the config of multiple terraform provider config, whose key is
// the provider name.
type TerraformRuntime map[string]GenericConfig
