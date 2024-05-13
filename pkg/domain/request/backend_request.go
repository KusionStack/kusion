package request

import v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"

// CreateBackendRequest represents the create request structure for
// backend.
type CreateBackendRequest struct {
	// Name is the name of the backend.
	Name string `json:"name" binding:"required"`
	// Description is a human-readable description of the backend.
	Description string `json:"description"`
	// BackendConfig is the configuration of the backend.
	BackendConfig v1.BackendConfig `json:"backendConfig"`
}

// UpdateBackendRequest represents the update request structure for
// backend.
type UpdateBackendRequest struct {
	// ID is the id of the backend.
	ID uint `json:"id" binding:"required"`
	// Name is the name of the backend.
	Name string `json:"name"`
	// Description is a human-readable description of the backend.
	Description string `json:"description"`
	// BackendConfig is the configuration of the backend.
	BackendConfig v1.BackendConfig `json:"backendConfig"`
}
