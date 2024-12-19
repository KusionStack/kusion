package request

import "net/http"

// CreateVariableSetRequest represents the create request structure
// for a variable in the variable set.
type CreateVariableSetRequest struct {
	// VariableKey is the access path for the variable.
	VariableKey string `json:"variableKey" binding:"required"`
	// Value is the value of the variable.
	Value string `json:"value" binding:"required"`
	// Type is the type of the variable.
	Type string `json:"type" binding:"required"`
	// Labels clarifies the scope of the variable.
	Labels map[string]string `json:"labels,omitempty"`
}

// UpdateVariableSetRequest represents the update request structure
// for a variable in the variable set.
type UpdateVariableSetRequest struct {
	// VariableKey is the access path for the variable.
	VariableKey string `json:"variableKey" binding:"required"`
	// Value is the value of the variable.
	Value string `json:"value" binding:"required"`
	// Type is the type of the variable.
	Type string `json:"type" binding:"required"`
	// Labels clarifies the scope of the variable.
	Labels map[string]string `json:"labels" binding:"required"`
	// Fqn is the fully qualified name of the variable.
	Fqn string `json:"fqn" binding:"required"`
}

// ListVariableSetRequest represents the list request structure
// for variables matched to the labels in the variable set.
type ListVariableSetRequest struct {
	// Labels clarifies the scope of the variables.
	Labels map[string]string `json:"labels" binding:"required"`
}

func (payload *CreateVariableSetRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateVariableSetRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *ListVariableSetRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
