package entity

import "errors"

const (
	PlainTextType  string = "PlainText"
	CipherTextType string = "CipherText"
)

// Variable represents the specific configuration code variable,
// which usually includes the global configurations for Terraform providers like
// api host, ak and sk.
type Variable struct {
	// VariableKey is the access path for the variable.
	VariableKey string `yaml:"variableKey,omitempty" json:"variableKey,omitempty"`
	// Value is the value of the variable.
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
	// Type is the type of the variable.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// Labels clarifies the scope of the variable.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Fqn is the fully qualified name of the variable.
	Fqn string `yaml:"fqn,omitempty" json:"fqn,omitempty"`
}

// VariableFilter represents the filter conditions to list variables.
type VariableFilter struct {
	Key        string
	Pagination *Pagination
}

// VariableListResult represents the result of listing variables.
type VariableListResult struct {
	Variables []*Variable
	Total     int
}

// Validate checks if the variable is valid.
// It returns an error if the variable is not valid.
func (v *Variable) Validate() error {
	if v == nil {
		return errors.New("variable is nil")
	}

	if v.VariableKey == "" {
		return errors.New("empty variable key")
	}

	if v.Type != PlainTextType && v.Type != CipherTextType {
		return errors.New("invalid variable type")
	}

	if v.Fqn == "" {
		return errors.New("empty fqn")
	}

	return nil
}
