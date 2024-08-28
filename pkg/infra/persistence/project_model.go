package persistence

import (
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// ProjectModel is a DO used to map the entity to the database.
type ProjectModel struct {
	gorm.Model
	Name           string `gorm:"index:unique_project,unique"`
	SourceID       uint
	Source         *SourceModel `gorm:"foreignKey:ID;references:SourceID"`
	OrganizationID uint
	Organization   *OrganizationModel `gorm:"foreignKey:ID;references:OrganizationID"`
	Path           string             `gorm:"index:unique_project,unique"`
	Description    string
	Labels         MultiString
	Owners         MultiString
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *ProjectModel) TableName() string {
	return "project"
}

// ToEntity converts the DO to an entity.
func (m *ProjectModel) ToEntity() (*entity.Project, error) {
	if m == nil {
		return nil, ErrProjectModelNil
	}

	var err error
	var sourceEntity *entity.Source
	var organizationEntity *entity.Organization
	if m.Source != nil {
		sourceEntity, err = m.Source.ToEntity()
		if err != nil {
			return nil, ErrFailedToConvertSourceToEntity
		}
	}
	if m.Organization != nil {
		organizationEntity, err = m.Organization.ToEntity()
		if err != nil {
			return nil, ErrFailedToConvertOrgToEntity
		}
	}

	return &entity.Project{
		ID:                m.ID,
		Name:              m.Name,
		Source:            sourceEntity,
		Organization:      organizationEntity,
		Path:              m.Path,
		Description:       m.Description,
		Labels:            []string(m.Labels),
		Owners:            []string(m.Owners),
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *ProjectModel) FromEntity(e *entity.Project) error {
	if m == nil {
		return ErrProjectModelNil
	}

	m.ID = e.ID
	m.Name = e.Name
	m.Description = e.Description
	m.Path = e.Path
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp
	if e.Source != nil {
		m.SourceID = e.Source.ID
		m.Source.FromEntity(e.Source)
	}
	if e.Organization != nil {
		m.OrganizationID = e.Organization.ID
		m.Organization.FromEntity(e.Organization)
	}

	return nil
}
