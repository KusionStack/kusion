package stack

import (
	"context"
	"errors"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"

	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *StackManager) ListStacks(ctx context.Context, filter *entity.StackFilter, sortOptions *entity.SortOptions) (*entity.StackListResult, error) {
	stackEntities, err := m.stackRepo.List(ctx, filter, sortOptions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}
	return stackEntities, nil
}

func (m *StackManager) GetStackByID(ctx context.Context, id uint) (*entity.Stack, error) {
	existingEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *StackManager) DeleteStackByID(ctx context.Context, id uint) error {
	err := m.stackRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}
	return nil
}

func (m *StackManager) UpdateStackByID(ctx context.Context, id uint, requestPayload request.UpdateStackRequest) (*entity.Stack, error) {
	// Convert request payload to domain model
	var requestEntity entity.Stack
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get project by id
	projectEntity, err := m.projectRepo.Get(ctx, requestPayload.ProjectID)
	if err != nil {
		return nil, err
	}
	requestEntity.Project = projectEntity

	// Get the existing stack by id
	updatedEntity, err := m.stackRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingStack
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update stack with repository
	err = m.stackRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *StackManager) CreateStack(ctx context.Context, requestPayload request.CreateStackRequest) (*entity.Stack, error) {
	logger := logutil.GetLogger(ctx)
	// Convert request payload to domain model
	var createdEntity entity.Stack
	err := copier.Copy(&createdEntity, &requestPayload)
	if err != nil {
		return nil, err
	}

	// Initialize
	createdEntity.CreationTimestamp = time.Now()
	createdEntity.UpdateTimestamp = time.Now()
	createdEntity.LastAppliedTimestamp = time.Unix(0, 0)

	var projectEntity *entity.Project
	// Get project entity
	if requestPayload.ProjectID != 0 {
		// If project id is provided, get project by id
		logger.Info("Project ID provided, getting project by ID...")
		projectEntity, err = m.projectRepo.Get(ctx, requestPayload.ProjectID)
		if err != nil {
			return nil, err
		}
		createdEntity.Project = projectEntity
	} else if requestPayload.ProjectName != "" {
		// Otherwise, get project by name
		logger.Info("Project name provided, getting project by name...")
		projectEntity, err = m.projectRepo.GetByName(ctx, requestPayload.ProjectName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, constant.ErrGettingNonExistingProject
			}
			return nil, err
		}
		createdEntity.Project = projectEntity
	} else {
		return nil, constant.ErrProjectNameOrIDRequired
	}

	// If explicit path is missing, build the stack path from project, stack name and optional cloud information
	if requestPayload.Path == "" {
		logger.Info("Path not explicitly provided, building stack path from project and stack name")
		stackPath, ok := buildValidStackPath(requestPayload, projectEntity)
		if !ok {
			return nil, constant.ErrInvalidStackPath
		}
		createdEntity.Path = stackPath
	}

	// The default state is UnSynced
	createdEntity.SyncState = constant.StackStateUnSynced
	// Create stack with repository
	err = m.stackRepo.Create(ctx, &createdEntity)
	if err != nil && err == gorm.ErrDuplicatedKey {
		return nil, constant.ErrStackAlreadyExists
	} else if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}
