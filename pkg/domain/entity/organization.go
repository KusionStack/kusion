package entity

import (
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// Organization represents the specific configuration organization
type Organization struct {
	// ID is the id of the organization.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the stack.
	Name string `yaml:"name" json:"name"`
	// DisplayName is the readability display name.
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	// Description is a human-readable description of the organization.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Labels are custom labels associated with the organization.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the organization.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// CreationTimestamp is the timestamp of the created for the organization.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the organization.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

// Validate checks if the organization is valid.
// It returns an error if the organization is not valid.
func (p *Organization) Validate() error {
	if p == nil {
		return constant.ErrOrgNil
	}

	if p.Name == "" {
		return constant.ErrOrgNameEmpty
	}

	if len(p.Owners) == 0 {
		return constant.ErrOrganizationOwnerNil
	}

	return nil
}
