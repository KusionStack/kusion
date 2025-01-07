package source

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

func (m *SourceManager) ListSources(ctx context.Context, filter *entity.SourceFilter) (*entity.SourceListResult, error) {
	sourceEntities, err := m.sourceRepo.List(ctx, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingSource
		}
		return nil, err
	}
	return sourceEntities, nil
}

func (m *SourceManager) GetSourceByID(ctx context.Context, id uint) (*entity.Source, error) {
	existingEntity, err := m.sourceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingSource
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *SourceManager) DeleteSourceByID(ctx context.Context, id uint) error {
	err := m.sourceRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingSource
		}
		return err
	}
	return nil
}

func (m *SourceManager) UpdateSourceByID(ctx context.Context, id uint, requestPayload request.UpdateSourceRequest) (*entity.Source, error) {
	// Convert request payload to domain model
	var requestEntity entity.Source
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Convert Remote string to URL
	remote, err := url.Parse(requestPayload.Remote)
	if err != nil {
		return nil, err
	}
	if remote.Scheme == "" {
		remote.Scheme = "https"
	}
	requestEntity.Remote = remote

	// Get the existing source by id
	updatedEntity, err := m.sourceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingSource
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update source with repository
	err = m.sourceRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *SourceManager) CreateSource(ctx context.Context, requestPayload request.CreateSourceRequest) (*entity.Source, error) {
	// Convert request payload to domain model
	var createdEntity entity.Source
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Convert Remote string to URL
	remote, err := url.Parse(requestPayload.Remote)
	if err != nil {
		return nil, err
	}
	if remote.Scheme == "" {
		remote.Scheme = "https"
	}

	createdEntity.Remote = remote

	// Create source with repository
	err = m.sourceRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (m *SourceManager) BuildSourceFilter(ctx context.Context, query *url.Values) (*entity.SourceFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building source filter...")

	filter := entity.SourceFilter{}

	sourceNameParam := query.Get("sourceName")
	if sourceNameParam != "" {
		filter.SourceName = sourceNameParam
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
