package persistence

import (
	"time"

	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// ResourceModel is a DO used to map the entity to the database.
type ResourceModel struct {
	gorm.Model
	StackID              uint
	Stack                *StackModel `gorm:"foreignKey:ID;references:StackID"`
	ResourceType         string
	ResourcePlane        string
	ResourceName         string
	KusionResourceID     string `gorm:"index:unique_kusion_resource_id,unique"`
	IAMResourceID        string
	CloudResourceID      string
	LastAppliedRevision  string
	LastAppliedTimestamp time.Time
	Status               string
	Attributes           map[string]any `gorm:"serializer:json" json:"attributes"`
	Extensions           map[string]any `gorm:"serializer:json" json:"extensions"`
	DependsOn            MultiString
	Provider             string
	Labels               MultiString
	Owners               MultiString
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *ResourceModel) TableName() string {
	return "resource"
}

// ToEntity converts the DO to an entity.
func (m *ResourceModel) ToEntity() (*entity.Resource, error) {
	if m == nil {
		return nil, ErrResourceModelNil
	}

	var err error
	var stackEntity *entity.Stack
	if m.Stack != nil {
		stackEntity, err = m.Stack.ToEntity()
		if err != nil {
			return nil, err
		}
	}

	return &entity.Resource{
		ID:                   m.ID,
		Stack:                stackEntity,
		ResourceType:         m.ResourceType,
		ResourcePlane:        m.ResourcePlane,
		ResourceName:         m.ResourceName,
		KusionResourceID:     m.KusionResourceID,
		IAMResourceID:        m.IAMResourceID,
		CloudResourceID:      m.CloudResourceID,
		LastAppliedRevision:  m.LastAppliedRevision,
		LastAppliedTimestamp: m.LastAppliedTimestamp,
		Status:               m.Status,
		Attributes:           m.Attributes,
		Extensions:           m.Extensions,
		DependsOn:            []string(m.DependsOn),
		Provider:             m.Provider,
		Labels:               []string(m.Labels),
		Owners:               []string(m.Owners),
		CreationTimestamp:    m.CreatedAt,
		UpdateTimestamp:      m.UpdatedAt,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *ResourceModel) FromEntity(e *entity.Resource) error {
	if m == nil {
		return ErrResourceModelNil
	}

	m.ID = e.ID
	m.ResourcePlane = e.ResourcePlane
	m.ResourceType = e.ResourceType
	m.ResourceName = e.ResourceName
	m.KusionResourceID = e.KusionResourceID
	m.IAMResourceID = e.IAMResourceID
	m.CloudResourceID = e.CloudResourceID
	m.LastAppliedRevision = e.LastAppliedRevision
	m.LastAppliedTimestamp = e.LastAppliedTimestamp
	m.Status = e.Status
	m.Attributes = e.Attributes
	m.Extensions = e.Extensions
	m.DependsOn = MultiString(e.DependsOn)
	m.Provider = e.Provider
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp
	if e.Stack != nil {
		m.StackID = e.Stack.ID
		m.Stack.FromEntity(e.Stack)
	}

	return nil
}

// FromEntity converts an entity to a DO.
func FromEntityList(entityList []*entity.Resource) ([]*ResourceModel, error) {
	dml := make([]*ResourceModel, 0)
	for _, entity := range entityList {
		var resourceModel ResourceModel
		err := resourceModel.FromEntity(entity)
		if err != nil {
			return nil, err
		}
		dml = append(dml, &resourceModel)
	}
	return dml, nil
}
