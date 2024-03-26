package persistence

import (
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// OrganizationModel is a DO used to map the entity to the database.
type OrganizationModel struct {
	gorm.Model
	Name        string `gorm:"index:unique_org,unique"`
	Description string
	Labels      MultiString
	Owners      MultiString
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *OrganizationModel) TableName() string {
	return "organization"
}

// ToEntity converts the DO to an entity.
func (m *OrganizationModel) ToEntity() (*entity.Organization, error) {
	if m == nil {
		return nil, ErrOrganizationModelNil
	}

	return &entity.Organization{
		ID:                m.ID,
		Name:              m.Name,
		Description:       m.Description,
		Labels:            []string(m.Labels),
		Owners:            []string(m.Owners),
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
	}, nil
}

// ToEntity converts the DO to an entity.
func (m *OrganizationModel) ToEntityWithSource(sourceEntity *entity.Source) (*entity.Organization, error) {
	if m == nil {
		return nil, ErrOrganizationModelNil
	}

	return &entity.Organization{
		ID:                m.ID,
		Name:              m.Name,
		Description:       m.Description,
		Labels:            []string(m.Labels),
		Owners:            []string(m.Owners),
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *OrganizationModel) FromEntity(e *entity.Organization) error {
	if m == nil {
		return ErrOrganizationModelNil
	}

	m.ID = e.ID
	m.Name = e.Name
	m.Description = e.Description
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp

	return nil
}
