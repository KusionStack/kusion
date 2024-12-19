package persistence

import (
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// VariableModel is a DO used to map the entity to the database.
type VariableModel struct {
	gorm.Model
	// VariableKey is the access path for the variable.
	VariableKey string
	// Value is the value of the variable.
	Value string
	// Type is the type of the variable.
	Type string
	// Labels clarifies the scope of the variable.
	Labels map[string]string `gorm:"serializer:json" json:"labels"`
	// Fqn is the fully qualified name of the variable.
	Fqn string `gorm:"index:unique_variable,unique"`
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (v *VariableModel) TableName() string {
	return "variable"
}

// ToEntity converts the DO to an entity.
func (v *VariableModel) ToEntity() (*entity.Variable, error) {
	if v == nil {
		return nil, ErrVariableModelNil
	}

	return &entity.Variable{
		VariableKey: v.VariableKey,
		Value:       v.Value,
		Type:        v.Type,
		Labels:      v.Labels,
		Fqn:         v.Fqn,
	}, nil
}

// FromEntity converts an entity to a DO.
func (v *VariableModel) FromEntity(e *entity.Variable) error {
	if v == nil {
		return ErrVariableModelNil
	}

	v.VariableKey = e.VariableKey
	v.Value = e.Value
	v.Type = e.Type
	v.Labels = e.Labels
	v.Fqn = e.Fqn

	return nil
}
