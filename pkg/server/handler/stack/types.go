package stack

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingStack  = errors.New("the stack does not exist")
	ErrUpdatingNonExistingStack = errors.New("the stack to update does not exist")
	ErrSourceNotFound           = errors.New("the specified source does not exist")
	ErrProjectNotFound          = errors.New("the specified project does not exist")
	ErrInvalidStacktID          = errors.New("the stack ID should be a uuid")
)

func NewHandler(
	orgRepository repository.OrganizationRepository,
	projectRepo repository.ProjectRepository,
	stackRepo repository.StackRepository,
	sourceRepo repository.SourceRepository,
) (*Handler, error) {
	return &Handler{
		orgRepository: orgRepository,
		stackRepo:     stackRepo,
		projectRepo:   projectRepo,
		sourceRepo:    sourceRepo,
	}, nil
}

type Handler struct {
	orgRepository repository.OrganizationRepository
	projectRepo   repository.ProjectRepository
	stackRepo     repository.StackRepository
	sourceRepo    repository.SourceRepository
}
