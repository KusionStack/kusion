package backend

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

func (m *BackendManager) ListBackends(ctx context.Context, filter *entity.BackendFilter, sortOptions *entity.SortOptions) (*entity.BackendListResult, error) {
	backendEntities, err := m.backendRepo.List(ctx, filter, sortOptions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingBackend
		}
		return nil, err
	}

	for i, entity := range backendEntities.Backends {
		entity, err := MaskBackendSensitiveData(entity)
		if err != nil {
			return nil, err
		}
		backendEntities.Backends[i] = entity
	}
	return backendEntities, nil
}

func (m *BackendManager) GetBackendByID(ctx context.Context, id uint) (*entity.Backend, error) {
	existingEntity, err := m.backendRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingBackend
		}
		return nil, err
	}

	existingEntity, err = MaskBackendSensitiveData(existingEntity)
	if err != nil {
		return nil, err
	}

	return existingEntity, nil
}

func (m *BackendManager) DeleteBackendByID(ctx context.Context, id uint) error {
	err := m.backendRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingBackend
		}
		return err
	}
	return nil
}

func (m *BackendManager) UpdateBackendByID(ctx context.Context, id uint, requestPayload request.UpdateBackendRequest) (*entity.Backend, error) {
	// Convert request payload to domain model
	var requestEntity entity.Backend
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing backend by id
	updatedEntity, err := m.backendRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingBackend
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update backend with repository
	err = m.backendRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}

	updatedEntity, err = MaskBackendSensitiveData(updatedEntity)
	if err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (m *BackendManager) CreateBackend(ctx context.Context, requestPayload request.CreateBackendRequest) (*entity.Backend, error) {
	// Convert request payload to domain model
	var createdEntity entity.Backend
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Create backend with repository
	err := m.backendRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}

	maskedEntity, err := MaskBackendSensitiveData(&createdEntity)
	if err != nil {
		return nil, err
	}

	return maskedEntity, nil
}

func (m *BackendManager) BuildBackendFilterAndSortOptions(ctx context.Context, query *url.Values) (*entity.BackendFilter, *entity.SortOptions, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building backend filter...")

	filter := entity.BackendFilter{}

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

	// Build sort options
	sortBy := query.Get("sortBy")
	sortBy, err := validateBackendSortOptions(sortBy)
	if err != nil {
		return nil, nil, err
	}
	SortOrderAscending, _ := strconv.ParseBool(query.Get("ascending"))
	backendSortOptions := &entity.SortOptions{
		Field:     sortBy,
		Ascending: SortOrderAscending,
	}

	return &filter, backendSortOptions, nil
}
