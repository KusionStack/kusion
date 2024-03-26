package project

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingProject  = errors.New("the project does not exist")
	ErrUpdatingNonExistingProject = errors.New("the project to update does not exist")
	ErrSourceNotFound             = errors.New("the specified source does not exist")
	ErrOrgNotFound                = errors.New("the specified org does not exist")
	ErrInvalidProjectID           = errors.New("the project ID should be a uuid")
)

func NewHandler(
	organizationRepo repository.OrganizationRepository,
	projectRepo repository.ProjectRepository,
	sourceRepo repository.SourceRepository,
) (*Handler, error) {
	return &Handler{
		organizationRepo: organizationRepo,
		projectRepo:      projectRepo,
		sourceRepo:       sourceRepo,
	}, nil
}

type Handler struct {
	organizationRepo repository.OrganizationRepository
	projectRepo      repository.ProjectRepository
	sourceRepo       repository.SourceRepository
}
