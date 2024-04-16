package persistence

import (
	"context"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/domain/repository"
)

// The sourceRepository type implements the repository.SourceRepository interface.
// If the sourceRepository type does not implement all the methods of the interface,
// the compiler will produce an error.
var _ repository.SourceRepository = &sourceRepository{}

// sourceRepository is a repository that stores sources in a gorm database.
type sourceRepository struct {
	// db is the underlying gorm database where sources are stored.
	db *gorm.DB
}

// NewSourceRepository creates a new source repository.
func NewSourceRepository(db *gorm.DB) repository.SourceRepository {
	return &sourceRepository{db: db}
}

// Create saves a source to the repository.
func (r *sourceRepository) Create(ctx context.Context, dataEntity *entity.Source) error {
	// r.db.AutoMigrate(&SourceModel{})
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel SourceModel
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

		// Map fresh record's data into Entity
		newEntity, err := dataModel.ToEntity()
		if err != nil {
			return err
		}
		*dataEntity = *newEntity

		return nil
	})
}

// Create saves a source to the repository.
func (r *sourceRepository) CreateOrUpdate(ctx context.Context, dataEntity *entity.Source) error {
	err := dataEntity.Validate()
	if err != nil {
		return err
	}

	// Map the data from Entity to DO
	var dataModel SourceModel
	err = dataModel.FromEntity(dataEntity)
	if err != nil {
		return err
	}

	// Check if the source exists and update it if it does
	if existingSource, err := r.GetByRemote(ctx, dataModel.Remote); err == nil && existingSource != nil {
		return r.Update(ctx, dataEntity)
	} else {
		return r.Create(ctx, dataEntity)
	}
}

// Delete removes a source from the repository.
func (r *sourceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dataModel SourceModel
		err := tx.WithContext(ctx).First(&dataModel, id).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Delete(&dataModel).Error
	})
}

// Update updates an existing source in the repository.
func (r *sourceRepository) Update(ctx context.Context, dataEntity *entity.Source) error {
	// Map the data from Entity to DO
	var dataModel SourceModel
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

// Get retrieves a source by its ID.
func (r *sourceRepository) Get(ctx context.Context, id uint) (*entity.Source, error) {
	var dataModel SourceModel
	err := r.db.WithContext(ctx).First(&dataModel, id).Error
	if err != nil {
		return nil, err
	}

	return dataModel.ToEntity()
}

// GetByRemote retrieves a source by its remote.
func (r *sourceRepository) GetByRemote(ctx context.Context, remote string) (*entity.Source, error) {
	var dataModel SourceModel
	err := r.db.WithContext(ctx).Where("remote = ?", remote).First(&dataModel).Error
	if err != nil {
		return nil, err
	}
	return dataModel.ToEntity()
}

// List retrieves all sources.
func (r *sourceRepository) List(ctx context.Context) ([]*entity.Source, error) {
	var dataModel []SourceModel
	sourceEntityList := make([]*entity.Source, 0)
	result := r.db.WithContext(ctx).Find(&dataModel)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, source := range dataModel {
		sourceEntity, err := source.ToEntity()
		if err != nil {
			return nil, err
		}
		sourceEntityList = append(sourceEntityList, sourceEntity)
	}
	return sourceEntityList, nil
}
