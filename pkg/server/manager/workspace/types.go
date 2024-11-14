package workspace

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingWorkspace  = errors.New("the workspace does not exist")
	ErrUpdatingNonExistingWorkspace = errors.New("the workspace to update does not exist")
	ErrInvalidWorkspaceID           = errors.New("the workspace ID should be a uuid")
	ErrBackendNotFound              = errors.New("the specified backend does not exist")
	ErrMsgModulesNotRegistered      = "the module%s %v %s not registered"
	ErrMsgModulesPathNotMatched     = "the oci path of the module%s %v %s not matched with the registered information"
)

type WorkspaceManager struct {
	workspaceRepo  repository.WorkspaceRepository
	backendRepo    repository.BackendRepository
	moduleRepo     repository.ModuleRepository
	defaultBackend entity.Backend
}

func NewWorkspaceManager(workspaceRepo repository.WorkspaceRepository,
	backendRepo repository.BackendRepository,
	moduleRepo repository.ModuleRepository,
	defaultBackend entity.Backend,
) *WorkspaceManager {
	return &WorkspaceManager{
		workspaceRepo:  workspaceRepo,
		backendRepo:    backendRepo,
		moduleRepo:     moduleRepo,
		defaultBackend: defaultBackend,
	}
}

func plural(count int) string {
	if count > 1 {
		return "s"
	}

	return ""
}

func verb(count int) string {
	if count > 1 {
		return "are"
	}

	return "is"
}
