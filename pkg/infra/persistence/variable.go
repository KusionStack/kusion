//nolint:dupl
package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

// The variableRepository type implements the repository.VariableRepository interface.
// If the variableRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.VariableRepository = &variableRepository{}

// variableRepository is a repository that stores variables in a gorm database.
type variableRepository struct {
	// db is the underlying gorm database where variables are stored.
	db *gorm.DB
}

// NewVariableRepository creates a new variable repository.
func NewVariableRepository(db *gorm.DB) repository.VariableRepository {
	return &variableRepository{db: db}
}

// Create saves a variable to the repository.
func (v *variableRepository) Create(ctx context.Context, dataEntity *entity.Variable) error {
	v.db.AutoMigrate(&VariableModel{})
	if err := dataEntity.Validate(); err != nil {
		return err
	}

	// Map the data from entity to DO.
	var dataModel VariableModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	return v.db.Transaction(func(tx *gorm.DB) error {
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

// Delete removes a variable from the repository.
func (v *variableRepository) Delete(ctx context.Context, fqn string) error {
	return v.db.Transaction(func(tx *gorm.DB) error {
		var dataModel VariableModel
		if err := tx.WithContext(ctx).
			Where("fqn = ?", fqn).First(&dataModel).Error; err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing variable in the repository.
func (v *variableRepository) Update(ctx context.Context, dataEntity *entity.Variable) error {
	// Map the data from Entity to DO.
	var dataModel VariableModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	if err := v.db.WithContext(ctx).
		Where("fqn = ?", dataModel.Fqn).Updates(&dataModel).Error; err != nil {
		return err
	}

	return nil
}

// GetByFqn retrieves a variable by its fqn.
func (v *variableRepository) GetByFqn(ctx context.Context, fqn string) (*entity.Variable, error) {
	var dataModel VariableModel
	if err := v.db.WithContext(ctx).Where("fqn = ?", fqn).First(&dataModel).Error; err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves all existing variables.
func (v *variableRepository) List(ctx context.Context, filter *entity.VariableFilter) (*entity.VariableListResult, error) {
	var dataModel []VariableModel
	variableEntityList := make([]*entity.Variable, 0)
	pattern, args := GetVariableQuery(filter)

	var result *gorm.DB
	if filter.Pagination.PageSize == 0 {
		// Fetch data without pagination.
		result = v.db.WithContext(ctx).Where(pattern, args...).Find(&dataModel)
	} else {
		// Fetch paginated data from result with offset and limit.
		offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
		result = v.db.WithContext(ctx).
			Where(pattern, args...).Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	for _, variable := range dataModel {
		variableEntity, err := variable.ToEntity()
		if err != nil {
			return nil, err
		}
		variableEntityList = append(variableEntityList, variableEntity)
	}

	// Get total rows.
	var totalRows int64
	result.Model(dataModel).Count(&totalRows)

	return &entity.VariableListResult{
		Variables: variableEntityList,
		Total:     int(totalRows),
	}, nil
}
