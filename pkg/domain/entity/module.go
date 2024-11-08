package entity

import (
	"fmt"
	"net/url"
)

// Module represents the specific configuration code module,
// which should be a specific instance of the Kusion Module provider.
type Module struct {
	// Name is the module name.
	Name string `yaml:"name" json:"name"`
	// URL is the module oci artifact registry URL.
	URL *url.URL `yaml:"url" json:"url"`
	// Description is a human-readable description of the module.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Owners is a list of owners for the module.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// Doc is the documentation URL of the module.
	Doc *url.URL `yaml:"doc,omitempty" json:"doc,omitempty"`
}

// Validate checks if the module is valid.
// It returns an error if the module is not valid.
func (m *Module) Validate() error {
	if m == nil {
		return fmt.Errorf("module is nil")
	}

	return nil
}
