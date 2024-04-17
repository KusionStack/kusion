package stack

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

const (
	Stdout      = "stdout"
	NoDiffFound = "All resources are reconciled. No diff found"
)

var (
	ErrGettingNonExistingStack         = errors.New("the stack does not exist")
	ErrUpdatingNonExistingStack        = errors.New("the stack to update does not exist")
	ErrSourceNotFound                  = errors.New("the specified source does not exist")
	ErrWorkspaceNotFound               = errors.New("the specified workspace does not exist")
	ErrProjectNotFound                 = errors.New("the specified project does not exist")
	ErrInvalidStackID                  = errors.New("the stack ID should be a uuid")
	ErrGettingNonExistingStateForStack = errors.New("can not find State in this stack")
	ErrNoManagedResourceToDestroy      = errors.New("no managed resources to destroy")
	ErrDryrunDestroy                   = errors.New("dryrun-mode is enabled, no resources will be destroyed")
)

type StackManager struct {
	stackRepo     repository.StackRepository
	projectRepo   repository.ProjectRepository
	workspaceRepo repository.WorkspaceRepository
}
