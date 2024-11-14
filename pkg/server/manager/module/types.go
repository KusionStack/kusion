package module

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingModule  = errors.New("the module does not exist")
	ErrUpdatingNonExistingModule = errors.New("the module to update does not exist")
	ErrEmptyModuleName           = errors.New("the module name should not be empty")
	ErrInvalidWorkspaceID        = errors.New("the workspace id is invalid")
)

type ModuleManager struct {
	moduleRepo    repository.ModuleRepository
	workspaceRepo repository.WorkspaceRepository
	backendRepo   repository.BackendRepository
}

func NewModuleManager(moduleRepo repository.ModuleRepository,
	workspaceRepo repository.WorkspaceRepository,
	backendRepo repository.BackendRepository,
) *ModuleManager {
	return &ModuleManager{
		moduleRepo:    moduleRepo,
		workspaceRepo: workspaceRepo,
		backendRepo:   backendRepo,
	}
}
