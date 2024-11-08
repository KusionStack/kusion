package module

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingModule  = errors.New("the module does not exist")
	ErrUpdatingNonExistingModule = errors.New("the module to update does not exist")
	ErrEmptyModuleName           = errors.New("the module name should not be empty")
)

type ModuleManager struct {
	moduleRepo repository.ModuleRepository
}

func NewModuleManager(moduleRepo repository.ModuleRepository) *ModuleManager {
	return &ModuleManager{
		moduleRepo: moduleRepo,
	}
}
