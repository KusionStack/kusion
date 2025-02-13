package request

import (
	"net/http"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// CreateVariableSetRequest represents the create request structure
// for a variable set.
type CreateVariableSetRequest struct {
	// Name is the name of the variable set.
	Name string `json:"name" binding:"required"`
	// Labels clarifies the scope of the variable set.
	Labels map[string]string `json:"labels" binding:"required"`
}

// UpdateVariableSetRequest represents the update request structure
// for a variable set.
type UpdateVariableSetRequest struct {
	// Name is the name of the variable set.
	Name string `json:"name" binding:"required"`
	// Labels clarifies the scope of the variable set.
	Labels map[string]string `json:"labels" binding:"required"`
}

func (payload *CreateVariableSetRequest) Validate() error {
	// Validate variable set name.
	if validName(payload.Name) {
		return constant.ErrInvalidVariableSetName
	}

	if len(payload.Labels) == 0 {
		return constant.ErrEmptyVariableSetLabels
	}

	return nil
}

func (payload *UpdateVariableSetRequest) Validate() error {
	// Validate variable set name.
	if payload.Name != "" && validName(payload.Name) {
		return constant.ErrInvalidVariableSetName
	}

	if len(payload.Labels) == 0 {
		return constant.ErrEmptyVariableSetLabels
	}

	return nil
}

func (payload *CreateVariableSetRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateVariableSetRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
