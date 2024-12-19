package variablelabels

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingVariable  = errors.New("the variable does not exist")
	ErrUpdatingNonExistingVariable = errors.New("the variable to update does not exist")
	ErrEmptyVariableKey            = errors.New("the variable key should not be empty")
	ErrEmptyVariableLabels         = errors.New("the variable labels should not be empty")
)

type VariableLabelsManager struct {
	variableLabelsRepo repository.VariableLabelsRepository
}

func NewVariableLabelsManager(
	variableLabelsRepo repository.VariableLabelsRepository,
) *VariableLabelsManager {
	return &VariableLabelsManager{
		variableLabelsRepo: variableLabelsRepo,
	}
}
