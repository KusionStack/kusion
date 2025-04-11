package variable

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingVariable  = errors.New("the variable does not exist")
	ErrUpdatingNonExistingVariable = errors.New("the variable to update does not exist")
	ErrEmptyVariableName           = errors.New("the variable name should not be empty")
)

type VariableManager struct {
	variableRepo repository.VariableRepository
}

func NewVariableManager(
	variableRepo repository.VariableRepository,
) *VariableManager {
	return &VariableManager{
		variableRepo: variableRepo,
	}
}
