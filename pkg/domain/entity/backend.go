package entity

import (
	"time"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
)

// Backend represents the specific configuration backend
type Backend struct {
	// ID is the id of the backend.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the backend.
	Name string `yaml:"name" json:"name"`
	// // Type is the type of the backend.
	// Type string `yaml:"type" json:"type"`
	// Backend is the configuration of the backend.
	BackendConfig v1.BackendConfig `yaml:"backendConfig" json:"backendConfig"`
	// Description is a human-readable description of the backend.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// CreationTimestamp is the timestamp of the created for the backend.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the backend.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

// Validate checks if the backend is valid.
// It returns an error if the backend is not valid.
func (w *Backend) Validate() error {
	if w == nil {
		return constant.ErrBackendNil
	}

	if w.Name == "" {
		return constant.ErrBackendNameEmpty
	}

	// if w.Type == "" {
	// 	return constant.ErrBackendTypeEmpty
	// }

	return nil
}
