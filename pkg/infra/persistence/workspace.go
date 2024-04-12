package persistence

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"

	"gorm.io/gorm"
)

// The workspaceRepository type implements the repository.WorkspaceRepository interface.
// If the workspaceRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.WorkspaceRepository = &workspaceRepository{}

// workspaceRepository is a repository that stores workspaces in a gorm database.
type workspaceRepository struct {
	// db is the underlying gorm database where workspaces are stored.
	db *gorm.DB
}

// NewWorkspaceRepository creates a new workspace repository.
func NewWorkspaceRepository(db *gorm.DB) repository.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

// Create saves a workspace to the repository.
func (r *workspaceRepository) Create(ctx context.Context, dataEntity *entity.Workspace) error {
	r.db.AutoMigrate(&WorkspaceModel{})
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel WorkspaceModel
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

// Delete removes a workspace from the repository.
func (r *workspaceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel WorkspaceModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Delete(&dataModel).Error
	})
}

// Update updates an existing workspace in the repository.
func (r *workspaceRepository) Update(ctx context.Context, dataEntity *entity.Workspace) error {
	// Map the data from Entity to DO
	var dataModel WorkspaceModel
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

// Get retrieves a workspace by its ID.
func (r *workspaceRepository) Get(ctx context.Context, id uint) (*entity.Workspace, error) {
	var dataModel WorkspaceModel
	err := r.db.WithContext(ctx).
		Preload("Backend").
		First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// GetByName retrieves a workspace by its name.
func (r *workspaceRepository) GetByName(ctx context.Context, name string) (*entity.Workspace, error) {
	var dataModel WorkspaceModel
	err := r.db.WithContext(ctx).
		Preload("Backend").
		Where("name = ?", name).First(&dataModel).Error
	if err != nil {
		return nil, err
	}
	return dataModel.ToEntity()
}

// List retrieves all workspaces.
func (r *workspaceRepository) List(ctx context.Context) ([]*entity.Workspace, error) {
	var dataModel []WorkspaceModel
	workspaceEntityList := make([]*entity.Workspace, 0)
	result := r.db.WithContext(ctx).
		Preload("Backend").
		Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, workspace := range dataModel {
		workspaceEntity, err := workspace.ToEntity()
		if err != nil {
			return nil, err
		}
		workspaceEntityList = append(workspaceEntityList, workspaceEntity)
	}
	return workspaceEntityList, nil
}
