package module

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"
	"kusionstack.io/kusion/pkg/server/manager/workspace"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *ModuleManager) CreateModule(ctx context.Context, requestPayload request.CreateModuleRequest) (*entity.Module, error) {
	// Convert request payload to the domain model.
	var createdEntity entity.Module
	if err := copier.Copy(&createdEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Parse remote string of `URL` and `Doc`.
	address, err := url.Parse(requestPayload.URL)
	if err != nil {
		return nil, err
	}
	if address.Scheme == "" {
		address.Scheme = "https"
	}
	createdEntity.URL = address

	if requestPayload.Doc != "" {
		doc, err := url.Parse(requestPayload.Doc)
		if err != nil {
			return nil, err
		}
		if doc.Scheme == "" {
			doc.Scheme = "https"
		}
		createdEntity.Doc = doc
	}

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

func (m *ModuleManager) ListModules(ctx context.Context, filter *entity.ModuleFilter) (*entity.ModuleListResult, error) {
	moduleEntities, err := m.moduleRepo.List(ctx, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingModule
		}

		return nil, err
	}

	return moduleEntities, nil
}

func (m *ModuleManager) ListModulesByWorkspaceID(ctx context.Context, workspaceID uint, filter *entity.ModuleFilter) (*entity.ModuleListResult, error) {
	// Get workspace entity by ID.
	existingEntity, err := m.workspaceRepo.Get(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workspace.ErrGettingNonExistingWorkspace
		}
		return nil, err
	}

	// Get backend by backend ID.
	backendEntity, err := m.backendRepo.Get(ctx, existingEntity.Backend.ID)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, workspace.ErrBackendNotFound
	} else if err != nil {
		return nil, err
	}

	// Generate backend from the backend entity.
	remoteBackend, err := workspace.NewBackendFromEntity(*backendEntity)
	if err != nil {
		return nil, err
	}

	// Get workspace storage from backend.
	wsStorage, err := remoteBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}

	// Get workspace config from storage.
	ws, err := wsStorage.Get(existingEntity.Name)
	if err != nil {
		return nil, err
	}

	// Get the modules in the workspace.
	moduleEntities := make([]*entity.ModuleWithVersion, 0, len(ws.Modules))
	for moduleName, moduleConfigs := range ws.Modules {
		// Skip if module name filter doesn't match
		if filter.ModuleName != "" && !strings.Contains(strings.ToLower(moduleName), strings.ToLower(filter.ModuleName)) {
			continue
		}

		moduleEntity, err := m.moduleRepo.Get(ctx, moduleName)
		if err != nil {
			return nil, err
		}

		moduleEntities = append(moduleEntities, &entity.ModuleWithVersion{
			Name:        moduleEntity.Name,
			URL:         moduleEntity.URL,
			Version:     moduleConfigs.Version,
			Description: moduleEntity.Description,
			Owners:      moduleEntity.Owners,
			Doc:         moduleEntity.Doc,
		})
	}

	// Calculate the pagination scope.
	// Note: we assume that the `Page` and `PageSize` here is always valid.
	start := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	end := filter.Pagination.Page*filter.Pagination.PageSize - 1
	if end > len(moduleEntities) {
		end = len(moduleEntities)
	}

	return &entity.ModuleListResult{
		ModulesWithVersion: moduleEntities[start:end],
		Total:              len(moduleEntities),
	}, nil
}

func (m *ModuleManager) BuildModuleFilter(ctx context.Context, query *url.Values) (*entity.ModuleFilter, error) {
	logger := logutil.GetLogger(ctx)
	logger.Info("Building module filter...")

	moduleNameParam := query.Get("moduleName")

	filter := entity.ModuleFilter{}
	if moduleNameParam != "" {
		filter.ModuleName = moduleNameParam
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
