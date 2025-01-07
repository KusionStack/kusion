package project

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *ProjectManager) ListProjects(ctx context.Context, filter *entity.ProjectFilter) (*entity.ProjectListResult, error) {
	projectEntities, err := m.projectRepo.List(ctx, filter)
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

	// Get the existing project by id
	updatedEntity, err := m.projectRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingProject
		}
		return nil, err
	}

	// Get source by id
	if requestPayload.SourceID == 0 {
		requestEntity.Source = updatedEntity.Source
	} else {
		// If sourceID is passed in, get source by id and update the project source
		sourceEntity, err := m.sourceRepo.Get(ctx, requestPayload.SourceID)
		if err != nil {
			return nil, err
		}
		requestEntity.Source = sourceEntity
	}

	// Get organization by id
	if requestPayload.OrganizationID == 0 {
		requestEntity.Organization = updatedEntity.Organization
	} else {
		// If orgID is passed in, get org by id and update the project organization
		organizationEntity, err := m.organizationRepo.Get(ctx, requestPayload.OrganizationID)
		if err != nil {
			return nil, err
		}
		requestEntity.Organization = organizationEntity
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update project with repository
	err = m.projectRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *ProjectManager) CreateProject(ctx context.Context, requestPayload request.CreateProjectRequest) (*entity.Project, error) {
	logger := logutil.GetLogger(ctx)
	// Convert request payload to domain model
	var createdEntity entity.Project
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// If sourceID is passed in, get source by id
	if requestPayload.SourceID != 0 {
		logger.Info("Source ID found in the request. Using the source ID...", "sourceID", requestPayload.SourceID)
		sourceEntity, err := m.sourceRepo.Get(ctx, requestPayload.SourceID)
		if err != nil {
			return nil, err
		}
		createdEntity.Source = sourceEntity
	} else {
		// if sourceID is not passed in, get source by default source remote
		sourceEntity, err := m.sourceRepo.GetByRemote(ctx, m.defaultSource.Remote.String())
		if err != nil && err == gorm.ErrRecordNotFound {
			// if a source with the default remote does not exist, create a new source
			logger.Info("Source not found, creating new source with default remote...", "remote", m.defaultSource.Remote)
			sourceEntity = &m.defaultSource
			if sourceEntity.Name == "" {
				sourceEntity.Name, err = GenerateDefaultSourceName(m.defaultSource.Remote.String())
				if err != nil {
					return nil, err
				}
			}
			err = m.sourceRepo.Create(ctx, sourceEntity)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			logger.Info("Source found with default remote. Using the source...", "source", sourceEntity.Remote.String())
		}
		createdEntity.Source = sourceEntity
	}

	// If orgID is passed in, get org by id
	if requestPayload.OrganizationID != 0 {
		logger.Info("Organization ID found in the request. Using the organization ID...", "organizationID", requestPayload.OrganizationID)
		organizationEntity, err := m.organizationRepo.Get(ctx, requestPayload.OrganizationID)
		if err != nil {
			return nil, err
		}
		createdEntity.Organization = organizationEntity
	} else {
		// if orgID is not passed in, get org by domain name
		organizationEntity, err := m.organizationRepo.GetByName(ctx, requestPayload.Domain)
		if err != nil && err == gorm.ErrRecordNotFound {
			// if an organization with the domain name does not exist, create a new organization
			logger.Info("Organization not found, creating new organization with domain name...", "domain", requestPayload.Domain)
			organizationEntity = &entity.Organization{
				Name:   requestPayload.Domain,
				Owners: []string{constant.DefaultOrgOwner},
			}
			err = m.organizationRepo.Create(ctx, organizationEntity)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			logger.Info("Organization found with domain name. Using the organization...", "organization", organizationEntity.Name)
		}
		createdEntity.Organization = organizationEntity
	}

	// Create project with repository
	err := m.projectRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (m *ProjectManager) BuildProjectFilter(ctx context.Context, query *url.Values) (*entity.ProjectFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building project filter...")

	filter := entity.ProjectFilter{}

	orgIDParam := query.Get("orgID")
	if orgIDParam != "" {
		orgID, err := strconv.Atoi(orgIDParam)
		if err != nil {
			return nil, constant.ErrInvalidOrganizationID
		}
		filter.OrgID = uint(orgID)
	}

	name := query.Get("name")
	if name != "" {
		filter.Name = name
	}

	// Set pagination parameters.
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = constant.CommonPageDefault
	}
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize <= 0 {
		pageSize = constant.CommonPageSizeDefault
	}
	filter.Pagination = &entity.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return &filter, nil
}
