package project

import (
	"errors"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

var (
	ErrGettingNonExistingProject  = errors.New("the project does not exist")
	ErrUpdatingNonExistingProject = errors.New("the project to update does not exist")
	ErrSourceNotFound             = errors.New("the specified source does not exist")
	ErrOrgNotFound                = errors.New("the specified org does not exist")
	ErrDefaultSourceRemoteInvalid = errors.New("the default source remote url is invalid")
)

type ProjectManager struct {
	projectRepo      repository.ProjectRepository
	organizationRepo repository.OrganizationRepository
	sourceRepo       repository.SourceRepository
	defaultSource    entity.Source
}

func NewProjectManager(projectRepo repository.ProjectRepository,
	organizationRepo repository.OrganizationRepository,
	sourceRepo repository.SourceRepository,
	defaultSource entity.Source,
) *ProjectManager {
	return &ProjectManager{
		projectRepo:      projectRepo,
		organizationRepo: organizationRepo,
		sourceRepo:       sourceRepo,
		defaultSource:    defaultSource,
	}
}
