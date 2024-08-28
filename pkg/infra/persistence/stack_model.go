package persistence

import (
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// StackModel is a DO used to map the entity to the database.
type StackModel struct {
	gorm.Model
	Name                  string `gorm:"index:unique_project,unique"`
	ProjectID             uint
	Project               *ProjectModel
	Description           string
	Type                  string
	Path                  string `gorm:"index:unique_project,unique"`
	DesiredVersion        string
	Labels                MultiString
	Owners                MultiString
	SyncState             string
	LastGeneratedRevision string
	LastPreviewedRevision string
	LastAppliedRevision   string
	LastAppliedTimestamp  time.Time
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *StackModel) TableName() string {
	return "stack"
}

// ToEntity converts the DO to an entity.
func (m *StackModel) ToEntity() (*entity.Stack, error) {
	if m == nil {
		return nil, ErrStackModelNil
	}

	stackState := constant.StackState(m.SyncState)

	projectEntity, err := m.Project.ToEntity()
	if err != nil {
		return nil, err
	}

	return &entity.Stack{
		ID:                    m.ID,
		Name:                  m.Name,
		Project:               projectEntity,
		Description:           m.Description,
		Type:                  m.Type,
		Path:                  m.Path,
		DesiredVersion:        m.DesiredVersion,
		Labels:                []string(m.Labels),
		Owners:                []string(m.Owners),
		SyncState:             stackState,
		LastGeneratedRevision: m.LastGeneratedRevision,
		LastPreviewedRevision: m.LastPreviewedRevision,
		LastAppliedRevision:   m.LastAppliedRevision,
		LastAppliedTimestamp:  m.LastAppliedTimestamp,
		CreationTimestamp:     m.CreatedAt,
		UpdateTimestamp:       m.UpdatedAt,
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
	m.Type = e.Type
	m.DesiredVersion = e.DesiredVersion
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.SyncState = string(e.SyncState)
	m.LastGeneratedRevision = e.LastGeneratedRevision
	m.LastPreviewedRevision = e.LastPreviewedRevision
	m.LastAppliedRevision = e.LastAppliedRevision
	m.LastAppliedTimestamp = e.LastAppliedTimestamp
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp
	// Convert the project to a DO
	if e.Project != nil {
		m.ProjectID = e.Project.ID
		m.Project.FromEntity(e.Project)
	}

	return nil
}
