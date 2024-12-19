package persistence

import (
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// VariableLabelsModel is a DO used to map the entity to the database.
type VariableLabelsModel struct {
	gorm.Model
	// VariableKey is the access path for the variable.
	VariableKey string `gorm:"index:unique_variable_labels,unique"`
	// Labels is the list of variable labels, which should be sorted
	// in ascending order of priority.
	Labels MultiString
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (vl *VariableLabelsModel) TableName() string {
	return "variable_labels"
}

// ToEntity converts the DO to an entity.
func (vl *VariableLabelsModel) ToEntity() (*entity.VariableLabels, error) {
	if vl == nil {
		return nil, ErrVariableLabelsModelNil
	}

	return &entity.VariableLabels{
		VariableKey: vl.VariableKey,
		Labels:      []string(vl.Labels),
	}, nil
}

// FromEntity converts an entity to a DO.
func (vl *VariableLabelsModel) FromEntity(e *entity.VariableLabels) error {
	if vl == nil {
		return ErrVariableLabelsModelNil
	}

	vl.VariableKey = e.VariableKey
	vl.Labels = MultiString(e.Labels)

	return nil
}
