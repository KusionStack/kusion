package persistence

import (
	"context"

	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"

	"gorm.io/gorm"
)

// The projectRepository type implements the repository.ProjectRepository interface.
// If the projectRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.ProjectRepository = &projectRepository{}

// projectRepository is a repository that stores projects in a gorm database.
type projectRepository struct {
	// db is the underlying gorm database where projects are stored.
	db *gorm.DB
}

// NewProjectRepository creates a new project repository.
func NewProjectRepository(db *gorm.DB) repository.ProjectRepository {
	return &projectRepository{db: db}
}

// Create saves a project to the repository.
func (r *projectRepository) Create(ctx context.Context, dataEntity *entity.Project) error {
	// r.db.AutoMigrate(&ProjectModel{})
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel ProjectModel
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

// Delete removes a project from the repository.
func (r *projectRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel ProjectModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Delete(&dataModel).Error
	})
}

// Update updates an existing project in the repository.
func (r *projectRepository) Update(ctx context.Context, dataEntity *entity.Project) error {
	// Map the data from Entity to DO
	var dataModel ProjectModel
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

// Get retrieves a project by its ID.
func (r *projectRepository) Get(ctx context.Context, id uint) (*entity.Project, error) {
	var dataModel ProjectModel
	err := r.db.WithContext(ctx).
		Preload("Source").
		Preload("Organization").
		First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// // GetByRemote retrieves a project by its remote.
// func (r *projectRepository) GetByRemote(ctx context.Context, remote string) (*entity.Project, error) {
// 	var dataModel ProjectModel
// 	err := r.db.WithContext(ctx).Where("remote = ?", remote).First(&dataModel).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return dataModel.ToEntity()
// }

// List retrieves all projects.
func (r *projectRepository) List(ctx context.Context) ([]*entity.Project, error) {
	var dataModel []ProjectModel
	projectEntityList := make([]*entity.Project, 0)
	result := r.db.WithContext(ctx).Preload("Source").Preload("Organization").Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, project := range dataModel {
		projectEntity, err := project.ToEntity()
		if err != nil {
			return nil, err
		}
		projectEntityList = append(projectEntityList, projectEntity)
	}
	return projectEntityList, nil
}
