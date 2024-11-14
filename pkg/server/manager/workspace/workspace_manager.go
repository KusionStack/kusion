package workspace

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
)

func (m *WorkspaceManager) ListWorkspaces(ctx context.Context, filter *entity.WorkspaceFilter) ([]*entity.Workspace, error) {
	workspaceEntities, err := m.workspaceRepo.List(ctx, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingWorkspace
		}
		return nil, err
	}
	return workspaceEntities, nil
}

func (m *WorkspaceManager) GetWorkspaceByID(ctx context.Context, id uint) (*entity.Workspace, error) {
	existingEntity, err := m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingWorkspace
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *WorkspaceManager) DeleteWorkspaceByID(ctx context.Context, id uint) error {
	err := m.workspaceRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingWorkspace
		}
		return err
	}
	return nil
}

func (m *WorkspaceManager) UpdateWorkspaceByID(ctx context.Context, id uint, requestPayload request.UpdateWorkspaceRequest) (*entity.Workspace, error) {
	// Convert request payload to domain model
	var requestEntity entity.Workspace
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing workspace by id
	updatedEntity, err := m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingWorkspace
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update workspace with repository
	err = m.workspaceRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *WorkspaceManager) CreateWorkspace(ctx context.Context, requestPayload request.CreateWorkspaceRequest) (*entity.Workspace, error) {
	// Convert request payload to domain model
	var createdEntity entity.Workspace
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get backend by id
	backendEntity, err := m.backendRepo.Get(ctx, requestPayload.BackendID)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, ErrBackendNotFound
	} else if err != nil {
		return nil, err
	}
	createdEntity.Backend = backendEntity

	// Generate backend from the backend entity.
	remoteBackend, err := NewBackendFromEntity(*backendEntity)
	if err != nil {
		return nil, err
	}

	// Get workspace storage from backend.
	wsStorage, err := remoteBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}

	// Create an initiated workspace config.
	if err = wsStorage.Create(&v1.Workspace{Name: createdEntity.Name}); err != nil {
		return nil, err
	}

	// Create workspace with repository
	err = m.workspaceRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}
