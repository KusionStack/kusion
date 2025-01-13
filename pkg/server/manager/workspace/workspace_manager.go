package workspace

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	backendmanager "kusionstack.io/kusion/pkg/server/manager/backend"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *WorkspaceManager) ListWorkspaces(ctx context.Context, filter *entity.WorkspaceFilter) (*entity.WorkspaceListResult, error) {
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

func (m *WorkspaceManager) DeleteWorkspaceByID(ctx context.Context, id uint) (err error) {
	// Get workspace by id
	existingEntity, err := m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingWorkspace
		}
		return err
	}

	// Get backend by id
	backendEntity, err := m.backendRepo.Get(ctx, existingEntity.Backend.ID)
	if err != nil && err == gorm.ErrRecordNotFound {
		return ErrBackendNotFound
	} else if err != nil {
		return err
	}

	// Generate backend from the backend entity.
	remoteBackend, err := NewBackendFromEntity(*backendEntity)
	if err != nil {
		return err
	}

	// Get workspace storage from backend.
	wsStorage, err := remoteBackend.WorkspaceStorage()
	if err != nil {
		return err
	}

	// Delete workspace with repository
	err = m.workspaceRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Rollback workspace if workspace storage deletion fails
	defer func() {
		if err != nil {
			_ = m.workspaceRepo.Create(ctx, existingEntity)
		}
	}()

	// Delete workspace storage
	err = wsStorage.Delete(existingEntity.Name)
	if err != nil {
		return err
	}

	return nil
}

func (m *WorkspaceManager) UpdateWorkspaceByID(ctx context.Context, id uint, requestPayload request.UpdateWorkspaceRequest) (updatedEntity *entity.Workspace, err error) {
	// Convert request payload to domain model
	var requestEntity entity.Workspace
	if err = copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing workspace by id
	updatedEntity, err = m.workspaceRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingWorkspace
		}
		return nil, err
	}
	beforeUpdatedEntity := *updatedEntity

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update workspace with repository
	err = m.workspaceRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}

	// Rollback workspace if rename workspace fails
	defer func() {
		if err != nil {
			_ = m.workspaceRepo.Update(ctx, &beforeUpdatedEntity)
		}
	}()

	if requestEntity.Name != "" && requestEntity.Name != beforeUpdatedEntity.Name {
		// Get backend by id
		backendEntity, err := m.backendRepo.Get(ctx, updatedEntity.Backend.ID)
		if err != nil && err == gorm.ErrRecordNotFound {
			return nil, ErrBackendNotFound
		} else if err != nil {
			return nil, err
		}

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

		// Rename workspace
		if err = wsStorage.RenameWorkspace(beforeUpdatedEntity.Name, requestEntity.Name); err != nil {
			return nil, err
		}
	}

	return updatedEntity, nil
}

func (m *WorkspaceManager) CreateWorkspace(ctx context.Context, requestPayload request.CreateWorkspaceRequest) (createdWorkspace *entity.Workspace, err error) {
	// Convert request payload to domain model
	var createdEntity entity.Workspace
	if err = copier.Copy(&createdEntity, &requestPayload); err != nil {
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

	// Ensure workspace storage is cleaned up if repository creation fails
	defer func() {
		if err != nil {
			_ = wsStorage.Delete(createdEntity.Name)
		}
	}()

	// Create workspace with repository
	err = m.workspaceRepo.Create(ctx, &createdEntity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (m *WorkspaceManager) BuildWorkspaceFilter(ctx context.Context, query *url.Values) (*entity.WorkspaceFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building workspace filter...")

	filter := entity.WorkspaceFilter{}

	backendIDParam := query.Get("backendID")
	if backendIDParam != "" {
		backendID, err := strconv.Atoi(backendIDParam)
		if err != nil {
			return nil, backendmanager.ErrInvalidBackendID
		}
		filter.BackendID = uint(backendID)
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
