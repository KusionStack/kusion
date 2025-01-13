//nolint:dupl
package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

// The moduleRepository type implements the repository.ModuleRepository interface.
// If the moduleRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.ModuleRepository = &moduleRepository{}

// moduleRepository is a repository that stores modules in a gorm database.
type moduleRepository struct {
	// db is the underlying gorm database where modules are stored.
	db *gorm.DB
}

// NewModuleRepository creates a new module repository.
func NewModuleRepository(db *gorm.DB) repository.ModuleRepository {
	return &moduleRepository{db: db}
}

// Create saves a module to the repository.
func (r *moduleRepository) Create(ctx context.Context, dataEntity *entity.Module) error {
	r.db.AutoMigrate(&ModuleModel{})
	if err := dataEntity.Validate(); err != nil {
		return err
	}

	// Map the data from Entity to DO.
	var dataModel ModuleModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create new record in the store.
		if err := tx.WithContext(ctx).Create(&dataModel).Error; err != nil {
			return err
		}

		// Map fresh record's data into Entity.
		newEntity, err := dataModel.ToEntity()
		if err != nil {
			return err
		}
		*dataEntity = *newEntity

		return nil
	})
}

// Delete removes a module from the repository.
func (r *moduleRepository) Delete(ctx context.Context, name string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel ModuleModel
		if err := tx.WithContext(ctx).Where("name = ?", name).First(&dataModel).Error; err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing module in the repository.
func (r *moduleRepository) Update(ctx context.Context, dataEntity *entity.Module) error {
	// Map the data from Entity to DO.
	var dataModel ModuleModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).
		Where("name = ?", dataModel.Name).Updates(&dataModel).Error; err != nil {
		return err
	}

	return nil
}

// Get retrieves a module by its name.
func (r *moduleRepository) Get(ctx context.Context, name string) (*entity.Module, error) {
	var dataModel ModuleModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&dataModel).Error; err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves all the modules.
func (r *moduleRepository) List(ctx context.Context, filter *entity.ModuleFilter) (*entity.ModuleListResult, error) {
	var dataModel []ModuleModel
	moduleEntityList := make([]*entity.Module, 0)
	pattern, args := GetModuleQuery(filter)
	searchResult := r.db.WithContext(ctx).Where(pattern, args...)

	// Get total rows
	var totalRows int64
	searchResult.Model(dataModel).Count(&totalRows)

	// Fetch paginated data from searchResult with offset and limit
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	result := searchResult.Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, module := range dataModel {
		moduleEntity, err := module.ToEntity()
		if err != nil {
			return nil, err
		}
		moduleEntityList = append(moduleEntityList, moduleEntity)
	}

	return &entity.ModuleListResult{
		Modules: moduleEntityList,
		Total:   int(totalRows),
	}, nil
}
