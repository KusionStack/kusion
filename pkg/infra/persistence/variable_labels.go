//nolint:dupl
package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

// The variableLabelsRepository type implements the repository.VariableLabelsRepository interface.
// If the variableLabelsRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.VariableLabelsRepository = &variableLabelsRepository{}

// variableLabelsRepository is a repository that stores variable labels in a gorm database.
type variableLabelsRepository struct {
	// db is the underlying gorm database where variable labels are stored.
	db *gorm.DB
}

// NewVariableLabelsRepository creates a new repository for variable labels.
func NewVariableLabelsRepository(db *gorm.DB) repository.VariableLabelsRepository {
	return &variableLabelsRepository{db: db}
}

// Create creates a new set of variable labels.
func (vl *variableLabelsRepository) Create(ctx context.Context, dataEntity *entity.VariableLabels) error {
	vl.db.AutoMigrate(&VariableLabelsModel{})
	if err := dataEntity.Validate(); err != nil {
		return err
	}

	// Map the data from entity to DO.
	var dataModel VariableLabelsModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	return vl.db.Transaction(func(tx *gorm.DB) error {
		// Create new record in the storage.
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

// Delete deletes a set of variable labels by its key.
func (vl *variableLabelsRepository) Delete(ctx context.Context, key string) error {
	return vl.db.Transaction(func(tx *gorm.DB) error {
		var dataModel VariableLabelsModel
		if err := tx.WithContext(ctx).
			Where("variable_key = ?", key).First(&dataModel).Error; err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing set of variable labels.
func (vl *variableLabelsRepository) Update(ctx context.Context, dataEntity *entity.VariableLabels) error {
	// Map the data from Entity to DO.
	var dataModel VariableLabelsModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	if err := vl.db.WithContext(ctx).
		Where("variable_key = ?", dataModel.VariableKey).Updates(&dataModel).Error; err != nil {
		return err
	}

	return nil
}

// GetByKey retrieves a set of variable labels by its key.
func (vl *variableLabelsRepository) GetByKey(ctx context.Context, key string) (*entity.VariableLabels, error) {
	var dataModel VariableLabelsModel
	if err := vl.db.WithContext(ctx).
		Where("variable_key = ?", key).First(&dataModel).Error; err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves all existing variable labels.
func (vl *variableLabelsRepository) List(ctx context.Context, filter *entity.VariableLabelsFilter) (*entity.VariableLabelsListResult, error) {
	var dataModel []VariableLabelsModel
	variableLabelsEntityList := make([]*entity.VariableLabels, 0)
	pattern, args := GetVariableLabelsQuery(filter)

	var result *gorm.DB
	if filter.Pagination.PageSize == 0 {
		// Fetch data without pagination.
		result = vl.db.WithContext(ctx).Where(pattern, args...).Find(&dataModel)
	} else {
		// Fetch paginated data from result with offset and limit.
		offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
		result = vl.db.WithContext(ctx).
			Where(pattern, args...).Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	for _, variableLabels := range dataModel {
		variableLabelsEntity, err := variableLabels.ToEntity()
		if err != nil {
			return nil, err
		}
		variableLabelsEntityList = append(variableLabelsEntityList, variableLabelsEntity)
	}

	// Get total rows.
	var totalRows int64
	result.Model(dataModel).Count(&totalRows)

	return &entity.VariableLabelsListResult{
		VariableLabels: variableLabelsEntityList,
		Total:          int(totalRows),
	}, nil
}
