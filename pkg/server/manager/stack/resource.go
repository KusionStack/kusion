package stack

import (
	"context"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func (m *StackManager) WriteResources(ctx context.Context, release *v1.Release, stack *entity.Stack, workspace, specID string) error {
	logger := logutil.GetLogger(ctx)
	logger.Info("Writing resources into database...")
	resourceEntitiesToInsert := []*entity.Resource{}

	if release.State != nil {
		for _, resource := range release.State.Resources {
			resourceEntity, err := convertV1ResourceToEntity(&resource)
			if err != nil {
				return err
			}
			resourceEntity.Stack = stack
			resourceEntity.LastAppliedRevision = specID
			resourceEntity.LastAppliedTimestamp = release.ModifiedTime
			resourceEntity.Attributes = resource.Attributes
			resourceEntity.Status = constant.StatusResourceApplied
			resourceEntity.Extensions = resource.Extensions
			resourceEntity.DependsOn = resource.DependsOn
			resourceEntity.ResourceURN = resourceURN(stack.Project.Name, stack.Name, workspace, resource.ID)
			resourceEntitiesToInsert = append(resourceEntitiesToInsert, resourceEntity)
		}
		if err := m.resourceRepo.Create(ctx, resourceEntitiesToInsert); err != nil {
			return err
		}
	}
	return nil
}

func (m *StackManager) MarkResourcesAsDeleted(ctx context.Context, release *v1.Release) error {
	logger := logutil.GetLogger(ctx)
	logger.Info("Marking resources as deleted in the database...")

	if release.State != nil {
		for _, resource := range release.State.Resources {
			resourceURN := resourceURN(release.Project, release.Stack, release.Workspace, resource.ID)
			resourceEntity, err := m.resourceRepo.GetByKusionResourceURN(ctx, resourceURN)
			if err != nil {
				return err
			}
			resourceEntity.LastAppliedTimestamp = release.ModifiedTime
			resourceEntity.Status = constant.StatusResourceDestroyed
			if err := m.resourceRepo.Update(ctx, resourceEntity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *StackManager) ReconcileResources(ctx context.Context, stackID uint, release *v1.Release) error {
	logger := logutil.GetLogger(ctx)
	logger.Info("Reconcile resources in the database for stack...")

	filter := &entity.ResourceFilter{
		StackID: stackID,
	}
	currentResources, err := m.resourceRepo.List(ctx, filter)
	if err != nil {
		return err
	}
	var resourceToBeDeleted []*entity.Resource

	for _, resource := range currentResources.Resources {
		if !isInRelease(release, resource.KusionResourceID, resource.Stack) {
			resource.LastAppliedTimestamp = release.ModifiedTime
			resource.Status = constant.StatusResourceDestroyed
			resourceToBeDeleted = append(resourceToBeDeleted, resource)
		}
	}

	if len(resourceToBeDeleted) > 0 {
		if err := m.resourceRepo.BatchDelete(ctx, resourceToBeDeleted); err != nil {
			return err
		}
	}

	return nil
}
