package persistence

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"

	"gorm.io/gorm"
)

// The stackRepository type implements the repository.StackRepository interface.
// If the stackRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.StackRepository = &stackRepository{}

// stackRepository is a repository that stores stacks in a gorm database.
type stackRepository struct {
	// db is the underlying gorm database where stacks are stored.
	db *gorm.DB
}

// NewStackRepository creates a new stack repository.
func NewStackRepository(db *gorm.DB) repository.StackRepository {
	return &stackRepository{db: db}
}

// Create saves a stack to the repository.
func (r *stackRepository) Create(ctx context.Context, dataEntity *entity.Stack) error {
	// r.db.AutoMigrate(&StackModel{})
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel StackModel
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

// Delete removes a stack from the repository.
func (r *stackRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel StackModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing stack in the repository.
func (r *stackRepository) Update(ctx context.Context, dataEntity *entity.Stack) error {
	// Map the data from Entity to DO
	var dataModel StackModel
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

// Get retrieves a stack by its ID.
func (r *stackRepository) Get(ctx context.Context, id uint) (*entity.Stack, error) {
	var dataModel StackModel
	err := r.db.WithContext(ctx).
		Preload("Project").Preload("Project.Organization").Preload("Project.Source").
		Joins("JOIN project ON project.id = stack.project_id").
		First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// List retrieves all stacks.
func (r *stackRepository) List(ctx context.Context, filter *entity.StackFilter) (*entity.StackListResult, error) {
	var dataModel []StackModel
	stackEntityList := make([]*entity.Stack, 0)
	pattern, args := GetStackQuery(filter)
	searchResult := r.db.WithContext(ctx).
		Preload("Project").Preload("Project.Organization").Preload("Project.Source").
		Joins("JOIN project ON project.id = stack.project_id").
		Where(pattern, args...)

	// Get total rows
	var totalRows int64
	searchResult.Model(dataModel).Count(&totalRows)

	// Fetch paginated data from searchResult with offset and limit
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	result := searchResult.Offset(offset).Limit(filter.Pagination.PageSize).Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, stack := range dataModel {
		stackEntity, err := stack.ToEntity()
		if err != nil {
			return nil, err
		}
		stackEntityList = append(stackEntityList, stackEntity)
	}
	return &entity.StackListResult{
		Stacks: stackEntityList,
		Total:  int(totalRows),
	}, nil
}
