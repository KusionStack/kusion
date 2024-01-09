package inputs

import (
	"fmt"
	"path/filepath"
	"strings"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

const (
	errInvalidProviderSource = "invalid provider source: %s"
	errEmptyProviderVersion  = "empty provider version"
)

const (
	RandomProvider   = "random"
	AWSProvider      = "aws"
	AlicloudProvider = "alicloud"
	defaultTFHost    = "registry.terraform.io"
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

// SetString sets the attributes into the provider object.
func (provider *Provider) SetString(providerURL string) error {
	// An example of the provider URL is shown below
	// registry.terraform.io/hashicorp/aws/5.0.1
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

// GetProviderURL returns the complete provider address from provider config in workspace.
func GetProviderURL(providerConfig *apiv1.ProviderConfig) (string, error) {
	if providerConfig.Version == "" {
		return "", fmt.Errorf(errEmptyProviderVersion)
	}

	// Conduct whether to use the default terraform provider registry host
	// according to the source of the provider config.
	// For example, "hashicorp/aws" means using the default tf provider registry,
	// while "registry.customized.io/hashicorp/aws" implies to use a customized registry host.
	attrs := strings.Split(providerConfig.Source, "/")
	if len(attrs) == 3 {
		return filepath.Join(providerConfig.Source, providerConfig.Version), nil
	} else if len(attrs) == 2 {
		return filepath.Join(defaultTFHost, providerConfig.Source, providerConfig.Version), nil
	}

	return "", fmt.Errorf(errInvalidProviderSource, providerConfig.Source)
}

// GetProviderRegion returns the region of the terraform provider.
func GetProviderRegion(providerConfig *apiv1.ProviderConfig) string {
	region, ok := providerConfig.GenericConfig["region"]
	if !ok {
		return ""
	}

	return region.(string)
}
