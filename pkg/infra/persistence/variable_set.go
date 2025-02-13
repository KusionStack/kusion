//nolint:dupl
package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

// The variableSetRepository type implements the repository.VariableSetRepository interface.
// If the variableSetRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.VariableSetRepository = &variableSetRepository{}

// variableSetRepository is a repository that stores variable sets in a gorm database.
type variableSetRepository struct {
	// db is the underlying gorm database where variable sets are stored.
	db *gorm.DB
}

// NewVariableSetRepository creates a new variableSet repository.
func NewVariableSetRepository(db *gorm.DB) repository.VariableSetRepository {
	return &variableSetRepository{db: db}
}

// Create saves a variable set to the repository.
func (vs *variableSetRepository) Create(ctx context.Context, dataEntity *entity.VariableSet) error {
	vs.db.AutoMigrate(&VariableSetModel{})
	if err := dataEntity.Validate(); err != nil {
		return err
	}

	// Map the data from entity to DO.
	var dataModel VariableSetModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	return vs.db.Transaction(func(tx *gorm.DB) error {
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

// Delete removes a variable set from the repository.
func (vs *variableSetRepository) Delete(ctx context.Context, name string) error {
	return vs.db.Transaction(func(tx *gorm.DB) error {
		var dataModel VariableSetModel
		if err := tx.WithContext(ctx).
			Where("name = ?", name).First(&dataModel).Error; err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing variable set in the repository.
func (vs *variableSetRepository) Update(ctx context.Context, dataEntity *entity.VariableSet) error {
	// Map the data from Entity to DO.
	var dataModel VariableSetModel
	if err := dataModel.FromEntity(dataEntity); err != nil {
		return err
	}

	if err := vs.db.WithContext(ctx).
		Where("name = ?", dataModel.Name).Updates(&dataModel).Error; err != nil {
		return err
	}

	return nil
}

// Get retrieves a variable set by its name.
func (vs *variableSetRepository) Get(ctx context.Context, name string) (*entity.VariableSet, error) {
	var dataModel VariableSetModel
	if err := vs.db.WithContext(ctx).Where("name = ?", name).First(&dataModel).Error; err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves existing variable sets with filter and sort options.
func (vs *variableSetRepository) List(ctx context.Context,
	filter *entity.VariableSetFilter, sortOptions *entity.SortOptions,
) (*entity.VariableSetListResult, error) {
	var dataModel []VariableSetModel
	variableSetEntityList := make([]*entity.VariableSet, 0)
	pattern, args := GetVariableSetQuery(filter)

	sortArgs := sortOptions.Field
	if sortOptions.Descending {
		sortArgs += " DESC"
	}

	searchResult := vs.db.WithContext(ctx).Order(sortArgs).Where(pattern, args...)

	// Get total rows.
	var totalRows int64
	searchResult.Model(dataModel).Count(&totalRows)

	// Set the page size as the maximum count limit of responses,
	// if `totalRows` is larger than `constant.CommonMaxResultLimit`.
	if totalRows > int64(constant.CommonMaxResultLimit) {
		filter.Pagination.Page = 1
		filter.Pagination.PageSize = constant.CommonMaxResultLimit
	} else if filter.FetchAll {
		// Set the page size as the `totalRows` if `filter.FetchAll` sets to true.
		filter.Pagination.Page = 1
		filter.Pagination.PageSize = int(totalRows)
	}

	// Fetched paginated data from searchResult with offset and limit.
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	result := searchResult.Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, variableSet := range dataModel {
		variableSetEntity, err := variableSet.ToEntity()
		if err != nil {
			return nil, err
		}
		variableSetEntityList = append(variableSetEntityList, variableSetEntity)
	}

	return &entity.VariableSetListResult{
		VariableSets: variableSetEntityList,
		Total:        int(totalRows),
	}, nil
}
