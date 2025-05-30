package entity

import "errors"

// VariableSet represents a set of the global configuration variables.
type VariableSet struct {
	// Name is the name of the variable set.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Labels clarifies the scope of the variable set.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// VariableSetFilter represents the filter conditions to list variable sets.
type VariableSetFilter struct {
	Name       string
	Pagination *Pagination
	FetchAll   bool
}

// VariableSetListResult represents the result of listing variable sets.
type VariableSetListResult struct {
	VariableSets []*VariableSet
	Total        int
}

// Validate checks if the variable set is valid.
func (vs *VariableSet) Validate() error {
	if vs == nil {
		return errors.New("variable set is nil")
	}

	if vs.Name == "" {
		return errors.New("empty variable set name")
	}

	if len(vs.Labels) == 0 {
		return errors.New("empty variable set labels")
	}

	return nil
}
