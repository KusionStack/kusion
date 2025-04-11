package persistence

import (
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// VariableModel is a DO used to map the entity to the database.
type VariableModel struct {
	gorm.Model
	// Name is the name of the variable.
	Name string `gorm:"index:unique_variable,unique"`
	// Value is the value of the variable.
	Value string
	// Type is the text type of the variable.
	Type entity.VariableType
	// VariableSet is the variable set to which the variable belongs.
	VariableSet string `gorm:"index:unique_variable,unique"`
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
		Name:        v.Name,
		Value:       v.Value,
		Type:        v.Type,
		VariableSet: v.VariableSet,
	}, nil
}

// FromEntity converts an entity to a DO.
func (v *VariableModel) FromEntity(e *entity.Variable) error {
	if v == nil {
		return ErrVariableModelNil
	}

	v.Name = e.Name
	v.Value = e.Value
	v.Type = e.Type
	v.VariableSet = e.VariableSet

	return nil
}
