package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

// The runRepository type implements the repository.RunRepository interface.
// If the runRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.RunRepository = &runRepository{}

// runRepository is a repository that stores runs in a gorm database.
type runRepository struct {
	// db is the underlying gorm database where runs are stored.
	db *gorm.DB
}

// NewRunRepository creates a new run repository.
func NewRunRepository(db *gorm.DB) repository.RunRepository {
	return &runRepository{db: db}
}

// Create saves a run to the repository.
func (r *runRepository) Create(ctx context.Context, dataEntity *entity.Run) error {
	// r.db.AutoMigrate(&RunModel{})
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel RunModel
	err = dataModel.FromEntity(dataEntity)
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.WithContext(ctx).Create(&dataModel).Error
		if err != nil {
			return err
		}

		dataEntity.ID = dataModel.ID

		return nil
	})
}

// Delete removes a run from the repository.
func (r *runRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel RunModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Unscoped().Delete(&dataModel).Error
	})
}

// Update updates an existing run in the repository.
func (r *runRepository) Update(ctx context.Context, dataEntity *entity.Run) error {
	// Map the data from Entity to DO
	var dataModel RunModel
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

// Get retrieves a run by its ID.
func (r *runRepository) Get(ctx context.Context, id uint) (*entity.Run, error) {
	var dataModel RunModel
	err := r.db.WithContext(ctx).
		Preload("Stack").Preload("Stack.Project").
		Joins("JOIN stack ON stack.id = run.stack_id").
		Joins("JOIN project ON project.id = stack.project_id").
		First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}
	return dataModel.ToEntity()
}

// List retrieves all runs.
func (r *runRepository) List(ctx context.Context, filter *entity.RunFilter) ([]*entity.Run, error) {
	var dataModel []RunModel
	runEntityList := make([]*entity.Run, 0)
	pattern, args := GetRunQuery(filter)
	result := r.db.WithContext(ctx).
		Preload("Stack").Preload("Stack.Project").
		Joins("JOIN stack ON stack.id = run.stack_id").
		Joins("JOIN project ON project.id = stack.project_id").
		Joins("JOIN workspace ON workspace.name = run.workspace").
		Where(pattern, args...).
		Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, run := range dataModel {
		runEntity, err := run.ToEntity()
		if err != nil {
			return nil, err
		}
		runEntityList = append(runEntityList, runEntity)
	}
	return runEntityList, nil
}
