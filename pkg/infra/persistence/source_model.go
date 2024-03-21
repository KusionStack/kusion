package persistence

import (
	"net/url"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// SourceModel is a DO used to map the entity to the database.
type SourceModel struct {
	gorm.Model
	// SourceProvider is the type of the source provider.
	SourceProvider string
	// Remote is the source URL, including scheme.
	Remote string
	// Description is a human-readable description of the source.
	Description string
	// Labels are custom labels associated with the source.
	Labels MultiString
	// Owners is a list of owners for the source.
	Owners MultiString
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *SourceModel) TableName() string {
	return "source"
}

// ToEntity converts the DO to an entity.
func (m *SourceModel) ToEntity() (*entity.Source, error) {
	if m == nil {
		return nil, ErrSourceModelNil
	}

	sourceProvider, err := constant.ParseSourceProviderType(m.SourceProvider)
	if err != nil {
		return nil, ErrFailedToGetSourceProviderType
	}

	var remote *url.URL
	if m.Remote == "local" {
		// convert string to url.URL
		remote, err = remote.Parse("local://file")
	} else {
		remote, err = url.Parse(m.Remote)
	}
	if err != nil {
		return nil, ErrFailedToGetSourceRemote
	}

	return &entity.Source{
		ID:                m.ID,
		SourceProvider:    sourceProvider,
		Remote:            remote,
		Description:       m.Description,
		Labels:            []string(m.Labels),
		Owners:            []string(m.Owners),
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *SourceModel) FromEntity(e *entity.Source) error {
	if m == nil {
		return ErrSourceModelNil
	}

	if e.Remote == nil || e.Remote.String() == "local://file" {
		m.Remote = "local"
	} else {
		m.Remote = e.Remote.String()
	}

	m.ID = e.ID
	m.SourceProvider = string(e.SourceProvider)
	//m.Remote = e.Remote.String()
	m.Description = e.Description
	m.Labels = MultiString(e.Labels)
	m.Owners = MultiString(e.Owners)
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp

	return nil
}
