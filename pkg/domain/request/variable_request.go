package request

import (
	"net/http"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// CreateVariableRequest represents the create request structure
// for a variable.
type CreateVariableRequest struct {
	// Name is the name of the variable.
	Name string `json:"name" binding:"required"`
	// Value is the value of the variable.
	Value string `json:"value"`
	// Type is the type of the variable.
	Type entity.VariableType `json:"type"`
	// VariableSet is the variable set to which the variable belongs.
	VariableSet string `json:"variableSet" binding:"required"`
}

// UpdateVariableRequest represents the update request structure
// for a variable.
type UpdateVariableRequest struct {
	// Name is the name of the variable.
	Name string `json:"name" binding:"required"`
	// Value is the value of the variable.
	Value string `json:"value"`
	// Type is the type of the variable.
	Type entity.VariableType `json:"type"`
	// VariableSet is the variable set to which the variable belongs.
	VariableSet string `json:"variableSet" binding:"required"`
}

func (payload *CreateVariableRequest) Validate() error {
	// Validate variable name.
	if validName(payload.Name) {
		return constant.ErrInvalidVariableName
	}

	// Validate variable set name. .
	if validName(payload.VariableSet) {
		return constant.ErrInvalidVariableSetName
	}

	// Validate variable type.
	if payload.Type != "" &&
		payload.Type != entity.PlainTextType && payload.Type != entity.CipherTextType {
		return constant.ErrInvalidVariableType
	}

	return nil
}

func (payload *UpdateVariableRequest) Validate() error {
	// Validate variable name.
	if validName(payload.Name) {
		return constant.ErrInvalidVariableName
	}

	// Validate variable set name. .
	if validName(payload.VariableSet) {
		return constant.ErrInvalidVariableSetName
	}

	// Validate variable type.
	if payload.Type != "" &&
		payload.Type != entity.PlainTextType && payload.Type != entity.CipherTextType {
		return constant.ErrInvalidVariableType
	}

	return nil
}

func (payload *CreateVariableRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateVariableRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
