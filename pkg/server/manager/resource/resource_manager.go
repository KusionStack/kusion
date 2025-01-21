package resource

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kusionstack.io/kusion/pkg/domain/entity"
)

func (m *ResourceManager) ListResources(ctx context.Context, filter *entity.ResourceFilter, sortOptions *entity.SortOptions) (*entity.ResourceListResult, error) {
	resourceEntities, err := m.resourceRepo.List(ctx, filter, sortOptions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingResource
		}
		return nil, err
	}
	return resourceEntities, nil
}

func (m *ResourceManager) GetResourceByID(ctx context.Context, id uint) (*entity.Resource, error) {
	existingEntity, err := m.resourceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingResource
		}
		return nil, err
	}
	return existingEntity, nil
}
