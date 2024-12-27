package persistence

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The resourceRepository type implements the repository.ResourceRepository interface.
// If the resourceRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.ResourceRepository = &resourceRepository{}

// resourceRepository is a repository that stores resources in a gorm database.
type resourceRepository struct {
	// db is the underlying gorm database where resources are stored.
	db *gorm.DB
}

// NewResourceRepository creates a new resource repository.
func NewResourceRepository(db *gorm.DB) repository.ResourceRepository {
	return &resourceRepository{db: db}
}

// Create saves a resource to the repository.
func (r *resourceRepository) Create(ctx context.Context, dataEntityList []*entity.Resource) error {
	// r.db.AutoMigrate(&ResourceModel{})
	for _, dataEntity := range dataEntityList {
		err := dataEntity.Validate()
		if err != nil {
			return err
		}
	}

	// Map the data from Entity to DO
	// var dataModel []ResourceModel
	dataModelList, err := FromEntityList(dataEntityList)
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create new record in the store
		err = tx.WithContext(ctx).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&dataModelList).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// Delete removes a resource from the repository.
func (r *resourceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel ResourceModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// BatchDelete removes a list of resources from the repository.
func (r *resourceRepository) BatchDelete(ctx context.Context, dataEntityList []*entity.Resource) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, dataEntity := range dataEntityList {
			err := dataEntity.Validate()
			if err != nil {
				return err
			}
		}

		dataModelList, err := FromEntityList(dataEntityList)
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModelList).Error
	})
}

// Update updates an existing resource in the repository.
func (r *resourceRepository) Update(ctx context.Context, dataEntity *entity.Resource) error {
	// Map the data from Entity to DO
	var dataModel ResourceModel
	err := dataModel.FromEntity(dataEntity)
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Updates(&dataModel).Error
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a resource by its ID.
func (r *resourceRepository) Get(ctx context.Context, id uint) (*entity.Resource, error) {
	var dataModel ResourceModel
	err := r.db.WithContext(ctx).
		Preload("Stack").Preload("Stack.Project").Preload("Stack.Project.Organization").Preload("Stack.Project.Source").
		Joins("JOIN stack ON stack.id = resource.stack_id").
		Joins("JOIN project ON project.id = stack.project_id").
		First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// GetByKusionResourceID retrieves a resource by its kusion resource id.
func (r *resourceRepository) GetByKusionResourceID(ctx context.Context, id string) (*entity.Resource, error) {
	var dataModel ResourceModel
	err := r.db.WithContext(ctx).
		Preload("Stack").Preload("Stack.Project").Preload("Stack.Project.Organization").Preload("Stack.Project.Source").
		Joins("JOIN stack ON stack.id = resource.stack_id").
		Joins("JOIN project ON project.id = stack.project_id").
		Where("kusion_resource_id = ?", id).
		First(&dataModel).Error
	if err != nil {
		return nil, err
	}
	return dataModel.ToEntity()
}

// GetByKusionResourceURN retrieves a resource by its kusion resource urn.
func (r *resourceRepository) GetByKusionResourceURN(ctx context.Context, id string) (*entity.Resource, error) {
	var dataModel ResourceModel
	err := r.db.WithContext(ctx).
		Preload("Stack").Preload("Stack.Project").Preload("Stack.Project.Organization").Preload("Stack.Project.Source").
		Joins("JOIN stack ON stack.id = resource.stack_id").
		Joins("JOIN project ON project.id = stack.project_id").
		Where("kusion_resource_urn = ?", id).
		First(&dataModel).Error
	if err != nil {
		return nil, err
	}
	return dataModel.ToEntity()
}

// List retrieves all resources.
func (r *resourceRepository) List(ctx context.Context, filter *entity.ResourceFilter) (*entity.ResourceListResult, error) {
	var dataModel []ResourceModel
	resourceEntityList := make([]*entity.Resource, 0)
	pattern, args := GetResourceQuery(filter)
	searchResult := r.db.WithContext(ctx).
		Preload("Stack").Preload("Stack.Project").Preload("Stack.Project.Organization").Preload("Stack.Project.Source").
		Joins("JOIN stack ON stack.id = resource.stack_id").
		Joins("JOIN project ON project.id = stack.project_id").
		Where(pattern, args...)

	// Get total rows
	var totalRows int64
	searchResult.Model(dataModel).Count(&totalRows)

	// Fetch paginated data from searchResult with offset and limit
	var result *gorm.DB
	if filter.Pagination != nil {
		offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
		result = searchResult.Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	} else {
		result = searchResult.Find(&dataModel)
	}
	if result.Error != nil {
		return nil, result.Error
	}

	for _, resource := range dataModel {
		resourceEntity, err := resource.ToEntity()
		if err != nil {
			return nil, err
		}
		resourceEntityList = append(resourceEntityList, resourceEntity)
	}
	return &entity.ResourceListResult{
		Resources: resourceEntityList,
		Total:     int(totalRows),
	}, nil
}
