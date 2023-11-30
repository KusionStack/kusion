package workspace

import (
	"errors"
	"reflect"
)

var (
	ErrEmptyKubeConfig                 = errors.New("empty kubeconfig")
	ErrEmptyTerraformProviderName      = errors.New("empty terraform provider name")
	ErrEmptyTerraformProviderConfig    = errors.New("empty terraform provider config")
	ErrNotExistTerraformProviderConfig = errors.New("not exist terraform provider config")
)

var ErrEmptyQueryTerraformProviderName = errors.New("empty query terraform provider name")

// RuntimeConfigs contains a set of runtime config.
type RuntimeConfigs struct {
	// Kubernetes contains the config to access a kubernetes cluster.
	Kubernetes KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`

	// Terraform contains the config of multiple terraform providers.
	Terraform TerraformConfig `yaml:"terraform,omitempty" json:"terraform,omitempty"`
}

// Validate is used to validate the RuntimeConfigs is valid or not.
func (r RuntimeConfigs) Validate() error {
	if !reflect.DeepEqual(r.Kubernetes, KubernetesConfig{}) {
		if err := r.Kubernetes.Validate(); err != nil {
			return err
		}
	}
	if len(r.Terraform) != 0 {
		if err := r.Terraform.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// KubernetesConfig contains config to access a kubernetes cluster.
type KubernetesConfig struct {
	// KubeConfig is the path of the kubeconfig file.
	KubeConfig string `yaml:"kubeConfig" json:"kubeConfig"`
}

// Validate is used to validate the KubernetesConfig is valid or not.
func (r KubernetesConfig) Validate() error {
	if r.KubeConfig == "" {
		return ErrEmptyKubeConfig
	}
	return nil
}

// TerraformConfig contains the config of multiple terraform provider config, whose key is
// the provider name.
type TerraformConfig map[string]GenericConfig

// Validate is used to validate the TerraformConfig is valid or not.
func (r TerraformConfig) Validate() error {
	for name, cfg := range r {
		if name == "" {
			return ErrEmptyTerraformProviderName
		}
		if len(cfg) == 0 {
			return ErrEmptyTerraformProviderConfig
		}
	}
	return nil
}

// GetProviderConfig is used to get a specified provider config.
func (r TerraformConfig) GetProviderConfig(providerName string) (GenericConfig, error) {
	if providerName == "" {
		return nil, ErrEmptyQueryTerraformProviderName
	}
	cfg, ok := r[providerName]
	if !ok {
		return nil, ErrNotExistTerraformProviderConfig
	}
	return cfg, nil
}
