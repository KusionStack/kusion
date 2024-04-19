package persistence

import (
	"context"
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// StackModel is a DO used to map the entity to the database.
type StackModel struct {
	gorm.Model
	Name string `gorm:"index:unique_project,unique"`
	// SourceID          uint
	// Source            *SourceModel
	ProjectID uint
	Project   *ProjectModel
	// OrganizationID    uint
	// Organization      *OrganizationModel
	Description       string
	Path              string `gorm:"index:unique_project,unique"`
	DesiredVersion    string
	Labels            MultiString
	Owners            MultiString
	SyncState         string
	LastSyncTimestamp time.Time
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *StackModel) TableName() string {
	return "stack"
}

// ToEntity converts the DO to an entity.
func (m *StackModel) ToEntity(ctx context.Context) (*entity.Stack, error) {
	if m == nil {
		return nil, ErrStackModelNil
	}

	stackState, err := constant.ParseStackState(m.SyncState)
	if err != nil {
		return nil, ErrFailedToGetStackState
	}

	// sourceEntity, err := m.Source.ToEntity()
	// if err != nil {
	// 	return nil, ErrFailedToConvertSourceToEntity
	// }

	// organizationEntity, err := m.Organization.ToEntity()
	// if err != nil {
	// 	return nil, ErrFailedToConvertSourceToEntity
	// }

	// projectEntity, err := m.Project.ToEntityWithSourceAndOrg(sourceEntity, organizationEntity)
	projectEntity, err := m.Project.ToEntityWithSourceAndOrg(nil, nil)
	if err != nil {
		return nil, ErrFailedToConvertProjectToEntity
	}

	return &entity.Stack{
		ID:   m.ID,
		Name: m.Name,
		// Source:            sourceEntity,
		Project: projectEntity,
		// Organization:      organizationEntity,
		Description:       m.Description,
		Path:              m.Path,
		DesiredVersion:    m.DesiredVersion,
		Labels:            []string(m.Labels),
		Owners:            []string(m.Owners),
		SyncState:         stackState,
		LastSyncTimestamp: m.LastSyncTimestamp,
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *StackModel) FromEntity(e *entity.Stack) error {
	if m == nil {
		return ErrStackModelNil
	}

	m.ID = e.ID
	m.Name = e.Name
	m.Description = e.Description
	m.Path = e.Path
	m.DesiredVersion = e.DesiredVersion
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.SyncState = string(e.SyncState)
	m.LastSyncTimestamp = e.LastSyncTimestamp
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp
	// Convert the source to a DO
	// if e.Source != nil {
	// 	m.SourceID = e.Source.ID
	// 	m.Source.FromEntity(e.Source)
	// }
	// Convert the project to a DO
	if e.Project != nil {
		m.ProjectID = e.Project.ID
		m.Project.FromEntity(e.Project)
	}
	// Convert the org to a DO
	// if e.Organization != nil {
	// 	m.OrganizationID = e.Organization.ID
	// 	m.Organization.FromEntity(e.Organization)
	// }

	return nil
}
