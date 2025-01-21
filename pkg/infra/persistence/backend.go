//nolint:dupl
package persistence

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"

	"gorm.io/gorm"
)

// The backendRepository type implements the repository.BackendRepository interface.
// If the backendRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.BackendRepository = &backendRepository{}

// backendRepository is a repository that stores backends in a gorm database.
type backendRepository struct {
	// db is the underlying gorm database where backends are stored.
	db *gorm.DB
}

// NewBackendRepository creates a new backend repository.
func NewBackendRepository(db *gorm.DB) repository.BackendRepository {
	return &backendRepository{db: db}
}

// Create saves a backend to the repository.
func (r *backendRepository) Create(ctx context.Context, dataEntity *entity.Backend) error {
	// r.db.AutoMigrate(&BackendModel{})
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel BackendModel
	err = dataModel.FromEntity(dataEntity)
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create new record in the store
		err = tx.WithContext(ctx).Create(&dataModel).Error
		if err != nil {
			return err
		}

		dataEntity.ID = dataModel.ID

		return nil
	})
}

// Delete removes a backend from the repository.
func (r *backendRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel BackendModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing backend in the repository.
func (r *backendRepository) Update(ctx context.Context, dataEntity *entity.Backend) error {
	// Map the data from Entity to DO
	var dataModel BackendModel
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

// Get retrieves a backend by its ID.
func (r *backendRepository) Get(ctx context.Context, id uint) (*entity.Backend, error) {
	var dataModel BackendModel
	err := r.db.WithContext(ctx).First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves all backends.
func (r *backendRepository) List(ctx context.Context, filter *entity.BackendFilter, sortOptions *entity.SortOptions) (*entity.BackendListResult, error) {
	var dataModel []BackendModel
	backendEntityList := make([]*entity.Backend, 0)

	sortArgs := sortOptions.Field
	if !sortOptions.Ascending {
		sortArgs += " DESC"
	}

	// Get total rows.
	var totalRows int64
	r.db.WithContext(ctx).Model(dataModel).Count(&totalRows)

	// Fetch paginated data with offset and limit.
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	result := r.db.WithContext(ctx).Order(sortArgs).Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, backend := range dataModel {
		backendEntity, err := backend.ToEntity()
		if err != nil {
			return nil, err
		}
		backendEntityList = append(backendEntityList, backendEntity)
	}
	return &entity.BackendListResult{
		Backends: backendEntityList,
		Total:    int(totalRows),
	}, nil
}
