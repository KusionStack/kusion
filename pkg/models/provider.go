package models

import (
	"fmt"
	"strings"
)

// Provider records the information of the Terraform provider
// used to provision cloud resources.
type Provider struct {
	// The complete provider address.
	URL string
	// The host of the provider registry.
	Host string
	// The namespace of the provider.
	Namespace string
	// The name of the provider.
	Name string
	// The version of the provider.
	Version string
}

// ProviderMeta records the metadata to declare a provider
// in Kusion's spec resources.
type ProviderMeta struct {
	// The region of provider resources.
	Region string `yaml:"region,omitempty" json:"region,omitempty"`
}

// SetString sets the attributes into the provider object.
func (provider *Provider) SetString(providerURL string) error {
	attrs := strings.Split(providerURL, "/")
	if len(attrs) != 4 {
		return fmt.Errorf("wrong provider url format: %s", providerURL)
	}

	provider.URL = providerURL
	provider.Host = attrs[0]
	provider.Namespace = attrs[1]
	provider.Name = attrs[2]
	provider.Version = attrs[3]

	return nil
}
