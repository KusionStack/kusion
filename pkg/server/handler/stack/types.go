package stack

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

const Stdout = "stdout"

var (
	ErrGettingNonExistingStack         = errors.New("the stack does not exist")
	ErrUpdatingNonExistingStack        = errors.New("the stack to update does not exist")
	ErrSourceNotFound                  = errors.New("the specified source does not exist")
	ErrWorkspaceNotFound               = errors.New("the specified workspace does not exist")
	ErrProjectNotFound                 = errors.New("the specified project does not exist")
	ErrInvalidStacktID                 = errors.New("the stack ID should be a uuid")
	ErrGettingNonExistingStateForStack = errors.New("can not find State in this stack")
)

func NewHandler(
	orgRepository repository.OrganizationRepository,
	projectRepo repository.ProjectRepository,
	stackRepo repository.StackRepository,
	sourceRepo repository.SourceRepository,
	workspaceRepo repository.WorkspaceRepository,
) (*Handler, error) {
	return &Handler{
		orgRepository: orgRepository,
		stackRepo:     stackRepo,
		projectRepo:   projectRepo,
		sourceRepo:    sourceRepo,
		workspaceRepo: workspaceRepo,
	}, nil
}

type Handler struct {
	orgRepository repository.OrganizationRepository
	projectRepo   repository.ProjectRepository
	stackRepo     repository.StackRepository
	sourceRepo    repository.SourceRepository
	workspaceRepo repository.WorkspaceRepository
}
