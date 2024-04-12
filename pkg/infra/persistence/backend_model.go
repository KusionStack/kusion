package persistence

import (
	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// type KusionBackend v1.BackendConfig

// func (b *KusionBackend) Value() (driver.Value, error) {
// 	if b == nil {
// 		return nil, nil
// 	}
// 	return json.Marshal(b)
// }

// func (b *KusionBackend) Scan(value interface{}) error {
// 	bytes, ok := value.([]byte)
// 	if !ok {
// 		return errors.New(fmt.Sprint("Failed to unmarshal KusionBackend value:", value))
// 	}

// 	return json.Unmarshal(bytes, b)
// }

// BackendModel is a DO used to map the entity to the database.
type BackendModel struct {
	gorm.Model
	Name          string `gorm:"index:unique_backend,unique"`
	Type          string `gorm:"index:unique_backend,unique"`
	Description   string
	Labels        MultiString
	Owners        MultiString
	BackendConfig v1.BackendConfig `gorm:"serializer:json" json:"backendConfig"`
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *BackendModel) TableName() string {
	return "backend"
}

// ToEntity converts the DO to an entity.
func (m *BackendModel) ToEntity() (*entity.Backend, error) {
	if m == nil {
		return nil, ErrBackendModelNil
	}

	return &entity.Backend{
		ID:   m.ID,
		Name: m.Name,
		//Type:              m.Type,
		Description:       m.Description,
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
		BackendConfig:     m.BackendConfig,
		// BackendConfig: v1.BackendConfig{
		// 	Type:    m.Type,
		// 	Configs: m.BackendConfig,
		// },
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *BackendModel) FromEntity(e *entity.Backend) error {
	if m == nil {
		return ErrBackendModelNil
	}

	m.ID = e.ID
	m.Name = e.Name
	//m.Type = e.Type
	m.Description = e.Description
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp
	//m.BackendConfig = e.BackendConfig.Configs
	m.BackendConfig = e.BackendConfig

	return nil
}
