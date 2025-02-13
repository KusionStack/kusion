package variableset

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingVariableSet  = errors.New("the variable set does not exist")
	ErrUpdatingNonExistingVariableSet = errors.New("the variable set to update does not exist")
	ErrEmptyVariableSetName           = errors.New("the variable set name should not be empty")
	ErrEmptyVariableSetLabels         = errors.New("the variable set labels should not be empty")
)

type VariableSetManager struct {
	variableSetRepo repository.VariableSetRepository
}

func NewVariableSetManager(
	variableSetRepo repository.VariableSetRepository,
) *VariableSetManager {
	return &VariableSetManager{
		variableSetRepo: variableSetRepo,
	}
}
