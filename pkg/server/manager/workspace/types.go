package workspace

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingWorkspace  = errors.New("the workspace does not exist")
	ErrUpdatingNonExistingWorkspace = errors.New("the workspace to update does not exist")
	ErrInvalidWorkspaceID           = errors.New("the workspace ID should be a uuid")
	ErrBackendNotFound              = errors.New("the specified backend does not exist")
)

type WorkspaceManager struct {
	workspaceRepo repository.WorkspaceRepository
	backendRepo   repository.BackendRepository
}

func NewWorkspaceManager(workspaceRepo repository.WorkspaceRepository, backendRepo repository.BackendRepository) *WorkspaceManager {
	return &WorkspaceManager{
		workspaceRepo: workspaceRepo,
		backendRepo:   backendRepo,
	}
}
