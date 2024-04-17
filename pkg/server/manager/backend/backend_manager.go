package backend

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

func (m *BackendManager) ListBackends(ctx context.Context) ([]*entity.Backend, error) {
	backendEntities, err := m.backendRepo.List(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingBackend
		}
		return nil, err
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
	return &createdEntity, nil
}
