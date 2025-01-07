package request

import (
	"net/http"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// CreateModuleRequest represents the create request structure for module.
type CreateModuleRequest struct {
	// Name is the module name.
	Name string `json:"name" binding:"required"`
	// URL is the module oci artifact registry URL.
	URL string `json:"url" binding:"required"`
	// Description is a human-readable description of the module.
	Description string `json:"description"`
	// Owners is a list of owners for the module.
	Owners []string `json:"owners"`
	// Doc is the documentation URL of the module.
	Doc string `json:"doc"`
}

// UpdateModuleRequest represents the update request structure for module.
type UpdateModuleRequest struct {
	// Name is the module name.
	Name string `json:"name" binding:"required"`
	// URL is the module oci artifact registry URL.
	URL string `json:"url"`
	// Description is a human-readable description of the module.
	Description string `json:"description"`
	// Owners is a list of owners for the module.
	Owners []string `json:"owners"`
	// Doc is the documentation URL of the module.
	Doc string `json:"doc"`
}

func (payload *CreateModuleRequest) Validate() error {
	// Validate module name
	if validName(payload.Name) {
		return constant.ErrInvalidModuleName
	}

	// Validate module URL and doc
	if err := validURL(payload.URL); err != nil {
		return err
	}

	if payload.Doc != "" {
		return validURL(payload.Doc)
	}

	return nil
}

func (payload *UpdateModuleRequest) Validate() error {
	// Validate module name
	if validName(payload.Name) {
		return constant.ErrInvalidModuleName
	}

	// Validate module URL and doc
	if payload.URL != "" {
		if err := validURL(payload.URL); err != nil {
			return err
		}
	}

	if payload.Doc != "" {
		return validURL(payload.Doc)
	}

	return nil
}

func (payload *CreateModuleRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateModuleRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
