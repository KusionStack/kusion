package organization

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *OrganizationManager) ListOrganizations(ctx context.Context, filter *entity.OrganizationFilter) (*entity.OrganizationListResult, error) {
	organizationEntities, err := m.organizationRepo.List(ctx, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingOrganization
		}
		return nil, err
	}
	return organizationEntities, nil
}

func (m *OrganizationManager) GetOrganizationByID(ctx context.Context, id uint) (*entity.Organization, error) {
	existingEntity, err := m.organizationRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingOrganization
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *OrganizationManager) DeleteOrganizationByID(ctx context.Context, id uint) error {
	err := m.organizationRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingOrganization
		}
		return err
	}
	return nil
}

func (m *OrganizationManager) UpdateOrganizationByID(ctx context.Context, id uint, requestPayload request.UpdateOrganizationRequest) (*entity.Organization, error) {
	// Convert request payload to domain model
	var requestEntity entity.Organization
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing organization by id
	updatedEntity, err := m.organizationRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingOrganization
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update organization with repository
	err = m.organizationRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (m *OrganizationManager) CreateOrganization(ctx context.Context, requestPayload request.CreateOrganizationRequest) (*entity.Organization, error) {
	// Convert request payload to domain model
	var createdEntity entity.Organization
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}
	// The default state is UnSynced
	createdEntity.CreationTimestamp = time.Now()
	createdEntity.UpdateTimestamp = time.Now()

	// Create organization with repository
	err := m.organizationRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (m *OrganizationManager) BuildOrganizationFilter(ctx context.Context, query *url.Values) (*entity.OrganizationFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building organization filter...")

	filter := entity.OrganizationFilter{}

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
