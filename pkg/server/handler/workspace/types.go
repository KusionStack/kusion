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

func NewHandler(
	workspaceRepo repository.WorkspaceRepository,
	backendRepo repository.BackendRepository,
) (*Handler, error) {
	return &Handler{
		workspaceRepo: workspaceRepo,
		backendRepo:   backendRepo,
	}, nil
}

type Handler struct {
	workspaceRepo repository.WorkspaceRepository
	backendRepo   repository.BackendRepository
}
