package stack

import (
	"context"
	"errors"
	"fmt"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/request"

	appmiddleware "kusionstack.io/kusion/pkg/server/middleware"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *StackManager) ListRuns(ctx context.Context, filter *entity.RunFilter, sortOptions *entity.SortOptions) (*entity.RunListResult, error) {
	runEntities, err := m.runRepo.List(ctx, filter, sortOptions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}
	return runEntities, nil
}

func (m *StackManager) GetRunByID(ctx context.Context, id uint) (*entity.Run, error) {
	existingEntity, err := m.runRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGettingNonExistingStack
		}
		return nil, err
	}
	return existingEntity, nil
}

func (m *StackManager) DeleteRunByID(ctx context.Context, id uint) error {
	err := m.runRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGettingNonExistingStack
		}
		return err
	}
	return nil
}

func (m *StackManager) UpdateRunByID(ctx context.Context, id uint, requestPayload request.UpdateRunRequest) (*entity.Run, error) {
	// Convert request payload to domain model
	var requestEntity entity.Run
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing stack by id
	updatedEntity, err := m.runRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingStack
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update stack with repository
	err = m.runRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (m *StackManager) CreateRun(ctx context.Context, requestPayload request.CreateRunRequest) (*entity.Run, error) {
	logger := logutil.GetLogger(ctx)
	// Convert request payload to domain model
	var createdEntity entity.Run
	err := copier.Copy(&createdEntity, &requestPayload)
	if err != nil {
		return nil, err
	}

	var stackEntity *entity.Stack
	if requestPayload.StackID != 0 {
		// If stack id is provided, get stack by id
		logger.Info("Stack ID provided, getting stack by ID...", "stackID", requestPayload.StackID)
		stackEntity, err = m.stackRepo.Get(ctx, requestPayload.StackID)
		if err != nil {
			return nil, err
		}
		createdEntity.Stack = stackEntity
	}

	logger.Info("Creating new run for stack and workspace", "stack", fmt.Sprint(createdEntity.Stack.ID), "workspace", createdEntity.Workspace)

	// The default status is InProgress
	createdEntity.Status = constant.RunStatusInProgress
	// Inject trace into run metadata
	traceID := appmiddleware.GetTraceID(ctx)
	createdEntity.Trace = traceID
	// Create run with repository
	err = m.runRepo.Create(ctx, &createdEntity)
	if err != nil && err == gorm.ErrDuplicatedKey {
		return nil, constant.ErrStackAlreadyExists
	} else if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (m *StackManager) UpdateRunResultAndStatusByID(ctx context.Context, id uint, requestPayload request.UpdateRunResultRequest) (*entity.Run, error) {
	// Convert request payload to domain model
	var requestEntity entity.Run
	if err := copier.Copy(&requestEntity, &requestPayload); err != nil {
		return nil, err
	}

	// Get the existing stack by id
	updatedEntity, err := m.runRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpdatingNonExistingStack
		}
		return nil, err
	}

	// Overwrite non-zero values in request entity to existing entity
	copier.CopyWithOption(updatedEntity, requestEntity, copier.Option{IgnoreEmpty: true})

	// Update stack with repository
	err = m.runRepo.Update(ctx, updatedEntity)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}
