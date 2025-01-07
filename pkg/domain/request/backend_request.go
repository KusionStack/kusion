package request

import (
	"net/http"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
)

// CreateBackendRequest represents the create request structure for
// backend.
type CreateBackendRequest struct {
	// Name is the name of the backend.
	Name string `json:"name" binding:"required"`
	// Description is a human-readable description of the backend.
	Description string `json:"description"`
	// BackendConfig is the configuration of the backend.
	BackendConfig v1.BackendConfig `json:"backendConfig" binding:"required"`
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

func (payload *CreateBackendRequest) Validate() error {
	// Validate backend name
	if validName(payload.Name) {
		return constant.ErrInvalidBackendName
	}

	// Validate backend type
	if payload.BackendConfig.Type == "" {
		return constant.ErrEmptyBackendType
	}

	if payload.BackendConfig.Type != v1.BackendTypeLocal &&
		payload.BackendConfig.Type != v1.BackendTypeOss &&
		payload.BackendConfig.Type != v1.BackendTypeS3 &&
		payload.BackendConfig.Type != v1.BackendTypeGoogle {
		return constant.ErrInvalidBackendType
	}

	return nil
}

func (payload *UpdateBackendRequest) Validate() error {
	if payload.Name != "" && validName(payload.Name) {
		return constant.ErrInvalidBackendName
	}

	if payload.BackendConfig.Type != "" &&
		payload.BackendConfig.Type != v1.BackendTypeLocal &&
		payload.BackendConfig.Type != v1.BackendTypeOss &&
		payload.BackendConfig.Type != v1.BackendTypeS3 &&
		payload.BackendConfig.Type != v1.BackendTypeGoogle {
		return constant.ErrInvalidBackendType
	}

	return nil
}

func (payload *CreateBackendRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateBackendRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
