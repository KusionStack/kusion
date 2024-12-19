package variable

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingVariable  = errors.New("the variable does not exist")
	ErrUpdatingNonExistingVariable = errors.New("the variable to update does not exist")
	ErrEmptyVariableKey            = errors.New("the variable key should not be empty")
	ErrEmptyVariableFqn            = errors.New("the variable fqn should not be empty")
)

type VariableManager struct {
	variableLabelsRepo repository.VariableLabelsRepository
	variableRepo       repository.VariableRepository
}

func NewVariableManager(
	variableRepo repository.VariableRepository,
	variableLabelsRepo repository.VariableLabelsRepository,
) *VariableManager {
	return &VariableManager{
		variableRepo:       variableRepo,
		variableLabelsRepo: variableLabelsRepo,
	}
}
