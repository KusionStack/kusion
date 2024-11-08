package module

import (
	"context"
	"errors"
	"net/url"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

func (m *ModuleManager) CreateModule(ctx context.Context, requestPayload request.CreateModuleRequest) (*entity.Module, error) {
	// Convert request payload to the domain model.
	var createdEntity entity.Module
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Parse remote string of `URL` and `Doc`.
	url, err := url.Parse(requestPayload.URL)
	if err != nil {
		return nil, err
	}
	createdEntity.URL = url

	doc, err := url.Parse(requestPayload.URL)
	if err != nil {
		return nil, err
	}
	createdEntity.Doc = doc

	// Create module with repository
	err = m.moduleRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (m *ModuleManager) DeleteModuleByName(ctx context.Context, name string) error {
	if err := m.moduleRepo.Delete(ctx, name); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingModule
		}
		return err
	}

	return nil
}

func (m *ModuleManager) UpdateModuleByName(ctx context.Context, name string, requestPayload request.UpdateModuleRequest) (*entity.Module, error) {
	// Convert request payload to domain model.
	var requestEntity entity.Module
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Parse remote string of `URL` and `Doc`.
	url, err := url.Parse(requestPayload.URL)
	if err != nil {
		return nil, err
	}
	requestEntity.URL = url

	doc, err := url.Parse(requestPayload.Doc)
	if err != nil {
		return nil, err
	}
	requestEntity.Doc = doc

	// Get the existing module by name.
	updatedEntity, err := m.moduleRepo.Get(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingModule
		}

		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity.
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update module with repository.
	if err = m.moduleRepo.Update(ctx, updatedEntity); err != nil {
		return nil, err
	}

	return updatedEntity, nil
}

func (m *ModuleManager) GetModuleByName(ctx context.Context, name string) (*entity.Module, error) {
	existingEntity, err := m.moduleRepo.Get(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingModule
		}

		return nil, err
	}

	return existingEntity, nil
}

func (m *ModuleManager) ListModules(ctx context.Context) ([]*entity.Module, error) {
	moduleEntities, err := m.moduleRepo.List(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingModule
		}

		return nil, err
	}

	return moduleEntities, nil
}
