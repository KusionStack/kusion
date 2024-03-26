package persistence

import (
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// WorkspaceModel is a DO used to map the entity to the database.
type WorkspaceModel struct {
	gorm.Model
	Name        string `gorm:"index:unique_workspace,unique"`
	Description string
	Labels      MultiString
	Owners      MultiString
	BackendID   uint
	Backend     *BackendModel `gorm:"foreignKey:ID;references:BackendID"`
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *WorkspaceModel) TableName() string {
	return "workspace"
}

// ToEntity converts the DO to an entity.
func (m *WorkspaceModel) ToEntity() (*entity.Workspace, error) {
	if m == nil {
		return nil, ErrWorkspaceModelNil
	}

	backendEntity, err := m.Backend.ToEntity()
	if err != nil {
		return nil, ErrFailedToConvertBackendToEntity
	}

	return &entity.Workspace{
		ID:                m.ID,
		Name:              m.Name,
		Description:       m.Description,
		Labels:            []string(m.Labels),
		Owners:            []string(m.Owners),
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
		Backend:           backendEntity,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *WorkspaceModel) FromEntity(e *entity.Workspace) error {
	if m == nil {
		return ErrWorkspaceModelNil
	}

	m.ID = e.ID
	m.Name = e.Name
	m.Description = e.Description
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp
	if e.Backend != nil {
		m.BackendID = e.Backend.ID
		m.Backend.FromEntity(e.Backend)
	}

	return nil
}
