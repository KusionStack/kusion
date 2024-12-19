package entity

import "errors"

// VariableLabels records the labels of the specific configuration code variable,
// and the labels are sorted in ascending order of priority.
type VariableLabels struct {
	// VariableKey is the access path for the variable.
	VariableKey string `yaml:"variableKey,omitempty" json:"variableKey,omitempty"`
	// Labels is the list of variable labels, which should be sorted
	// in ascending order of priority.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// VariableLabelsFilter represents the filter conditions to list variable labels.
type VariableLabelsFilter struct {
	Labels     []string
	Pagination *Pagination
}

// VariableLabelsListResult represents the result of listing variable labels.
type VariableLabelsListResult struct {
	VariableLabels []*VariableLabels
	Total          int
}

// Validate checks if the variable labels are valid.
// It returns an error if the variable labels are not valid.
func (vl *VariableLabels) Validate() error {
	if vl == nil {
		return errors.New("nil variable labels")
	}

	if vl.VariableKey == "" {
		return errors.New("empty key for variable labels")
	}

	if len(vl.Labels) == 0 {
		return errors.New("empty variable labels")
	}

	return nil
}
