package project

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/handler"
)

func (m *ProjectManager) ListProjects(ctx context.Context) ([]*entity.Project, error) {
	projectEntities, err := m.projectRepo.List(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingProject
		}
		return nil, err
	}
	return projectEntities, nil
}

func (m *ProjectManager) GetProjectByID(ctx context.Context, id uint) (*entity.Project, error) {
	existingEntity, err := m.projectRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingProject
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *ProjectManager) DeleteProjectByID(ctx context.Context, id uint) error {
	err := m.projectRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingProject
		}
		return err
	}
	return nil
}

func (m *ProjectManager) UpdateProjectByID(ctx context.Context, id uint, requestPayload request.UpdateProjectRequest) (*entity.Project, error) {
	// Convert request payload to domain model
	var requestEntity entity.Project
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get source by id
	sourceEntity, err := handler.GetSourceByID(ctx, m.sourceRepo, requestPayload.SourceID)
	if err != nil {
		return nil, err
	}
	requestEntity.Source = sourceEntity

	// Get organization by id
	organizationEntity, err := handler.GetOrganizationByID(ctx, m.organizationRepo, requestPayload.OrganizationID)
	if err != nil {
		return nil, err
	}
	requestEntity.Organization = organizationEntity

	// Get the existing project by id
	updatedEntity, err := m.projectRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingProject
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})
	// fmt.Printf("updatedEntity.Source: %v; updatedEntity.Organization: %v", updatedEntity.Source, updatedEntity.Organization)

	// Update project with repository
	err = m.projectRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *ProjectManager) CreateProject(ctx context.Context, requestPayload request.CreateProjectRequest) (*entity.Project, error) {
	// Convert request payload to domain model
	var createdEntity entity.Project
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get source by id
	sourceEntity, err := m.sourceRepo.Get(ctx, requestPayload.SourceID)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, ErrSourceNotFound
	} else if err != nil {
		return nil, err
	}
	createdEntity.Source = sourceEntity

	// Get org by id
	organizationEntity, err := m.organizationRepo.Get(ctx, requestPayload.OrganizationID)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, ErrOrgNotFound
	} else if err != nil {
		return nil, err
	}
	createdEntity.Organization = organizationEntity

	// Create project with repository
	err = m.projectRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}
