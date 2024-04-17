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

type ProjectManager struct {
	projectRepo      repository.ProjectRepository
	organizationRepo repository.OrganizationRepository
	sourceRepo       repository.SourceRepository
}

func NewProjectManager(projectRepo repository.ProjectRepository, organizationRepo repository.OrganizationRepository, sourceRepo repository.SourceRepository) *ProjectManager {
	return &ProjectManager{
		projectRepo:      projectRepo,
		organizationRepo: organizationRepo,
		sourceRepo:       sourceRepo,
	}
}
