package persistence

import (
	"net/url"

	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// ModuleModel is a DO used to map the entity to the database.
type ModuleModel struct {
	gorm.Model
	// Name is the module name.
	Name string `gorm:"index:unique_module,unique"`
	// URL is the module oci artifact registry URL.
	URL string
	// Description is a human-readable description of the module.
	Description string
	// Owners is a list of owners for the module.
	Owners MultiString
	// Doc is the documentation URL of the module.
	Doc string
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *ModuleModel) TableName() string {
	return "module"
}

// ToEntity converts the DO to an entity.
func (m *ModuleModel) ToEntity() (*entity.Module, error) {
	if m == nil {
		return nil, ErrModuleModelNil
	}

	url, err := url.Parse(m.URL)
	if err != nil {
		return nil, ErrFailedToGetModuleRemote
	}

	doc, err := url.Parse(m.Doc)
	if err != nil {
		return nil, ErrFailedToGetModuleDocRemote
	}

	return &entity.Module{
		Name:        m.Name,
		URL:         url,
		Description: m.Description,
		Owners:      []string(m.Owners),
		Doc:         doc,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *ModuleModel) FromEntity(e *entity.Module) error {
	if m == nil {
		return ErrModuleModelNil
	}

	if e.URL == nil {
		m.URL = ""
	} else {
		m.URL = e.URL.String()
	}

	if e.Doc == nil {
		m.Doc = ""
	} else {
		m.Doc = e.Doc.String()
	}

	m.Name = e.Name
	m.Description = e.Description
	m.Owners = MultiString(e.Owners)

	return nil
}
