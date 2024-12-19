package request

import "net/http"

// CreateVariableLabelsRequest represents the create request structure
// for variable labels.
type CreateVariableLabelsRequest struct {
	// VariableKey is the access path for the variable.
	VariableKey string `json:"variableKey" binding:"required"`
	// Labels is the list of variable labels, which should be sorted
	// in ascending order of priority.
	Labels []string `json:"labels" binding:"required"`
}

// UpdateVariableLabelsRequest represents the update request structure
// for variable labels.
type UpdateVariableLabelsRequest struct {
	// VariableKey is the access path for the variable.
	VariableKey string `json:"variableKey" binding:"required"`
	// Labels is the list of variable labels, which should be sorted
	// in ascending order of priority.
	Labels []string `json:"labels" binding:"required"`
}

func (payload *CreateVariableLabelsRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateVariableLabelsRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
