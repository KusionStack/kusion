package persistence

import (
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// VariableSetModel is a DO used to map the entity to the database.
type VariableSetModel struct {
	gorm.Model
	// Name is the name of the variable set.
	Name string `gorm:"index:unique_variable_set,unique"`
	// Labels clarifies the scope of the variable set.
	Labels map[string]string `gorm:"serializer:json" json:"labels"`
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (vs *VariableSetModel) TableName() string {
	return "variable_set"
}

// ToEntity converts the DO to an entity.
func (vs *VariableSetModel) ToEntity() (*entity.VariableSet, error) {
	if vs == nil {
		return nil, ErrVariableSetModelNil
	}

	return &entity.VariableSet{
		Name:   vs.Name,
		Labels: vs.Labels,
	}, nil
}

// FromEntity converts an entity to a DO.
func (vs *VariableSetModel) FromEntity(e *entity.VariableSet) error {
	if vs == nil {
		return ErrVariableSetModelNil
	}

	vs.Name = e.Name
	vs.Labels = e.Labels

	return nil
}
