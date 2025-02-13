package entity

import "errors"

type VariableType string

const (
	PlainTextType  VariableType = "PlainText"
	CipherTextType VariableType = "CipherText"
)

// Variable represents a specific configuration code variable,
// which usually includes the global configuration for Terraform providers like
// api host, access key and secret key.
type Variable struct {
	// Name is the name of the variable.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Value is the value of the variable.
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
	// Type is the text type of the variable.
	Type VariableType `yaml:"type,omitempty" json:"type,omitempty"`
	// VariableSet is the variable set to which the variable belongs.
	VariableSet string `yaml:"variableSet,omitempty" json:"variableSet,omitempty"`
}

// VariableFilter represents the filter conditions to list variables.
type VariableFilter struct {
	Name        string
	VariableSet string
	Pagination  *Pagination
	FetchAll    bool
}

// VariableListResult represents the result of listing variables.
type VariableListResult struct {
	Variables []*Variable
	Total     int
}

// Validate checks if the variable is valid.
func (v *Variable) Validate() error {
	if v == nil {
		return errors.New("variable is nil")
	}

	if v.Name == "" {
		return errors.New("empty variable name")
	}

	if v.Type != PlainTextType && v.Type != CipherTextType {
		return errors.New("invalid variable type")
	}

	if v.VariableSet == "" {
		return errors.New("empty variable set name")
	}

	return nil
}
