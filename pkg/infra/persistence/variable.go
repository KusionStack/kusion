//nolint:dupl
package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
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

	// Map the data from entity to Do.
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
func (v *variableRepository) Delete(ctx context.Context, name, variableSet string) error {
	return v.db.Transaction(func(tx *gorm.DB) error {
		var dataModel VariableModel
		if err := tx.WithContext(ctx).
			Where("name = ?", name).Where("variable_set = ?", variableSet).First(&dataModel).Error; err != nil {
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
		Where("name = ?", dataModel.Name).Where("variable_set = ?", dataModel.VariableSet).Updates(&dataModel).Error; err != nil {
		return err
	}

	return nil
}

// Get retrieves a variable by its name and the variable set it belongs to.
func (v *variableRepository) Get(ctx context.Context, name, variableSet string) (*entity.Variable, error) {
	var dataModel VariableModel
	if err := v.db.WithContext(ctx).
		Where("name = ?", name).Where("variable_set = ?", variableSet).First(&dataModel).Error; err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves existing variables with filter and sort options.
func (v *variableRepository) List(ctx context.Context,
	filter *entity.VariableFilter, sortOptions *entity.SortOptions,
) (*entity.VariableListResult, error) {
	var dataModel []VariableModel
	variableEntityList := make([]*entity.Variable, 0)
	pattern, args := GetVariableQuery(filter)

	sortArgs := sortOptions.Field
	if sortOptions.Descending {
		sortArgs += " DESC"
	}

	searchResult := v.db.WithContext(ctx).Order(sortArgs).Where(pattern, args...)

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

	for _, variable := range dataModel {
		variableEntity, err := variable.ToEntity()
		if err != nil {
			return nil, err
		}
		variableEntityList = append(variableEntityList, variableEntity)
	}

	return &entity.VariableListResult{
		Variables: variableEntityList,
		Total:     int(totalRows),
	}, nil
}
