package entity

import (
	"time"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
)

// Project represents the specific configuration project
type Project struct {
	// ID is the id of the project.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the project.
	Name string `yaml:"name" json:"name"`
	// DisplayName is the human-readable display name.
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	// Source is the configuration source associated with the project.
	Source *Source `yaml:"source" json:"source"`
	// Organization is the configuration source associated with the project.
	Organization *Organization `yaml:"organization" json:"organization"`
	// Description is a human-readable description of the project.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Path is the relative path of the project within the sourcs.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Labels are custom labels associated with the project.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the project.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// CreationTimestamp is the timestamp of the created for the project.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the project.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

// Validate checks if the project is valid.
// It returns an error if the project is not valid.
func (p *Project) Validate() error {
	if p == nil {
		return constant.ErrProjectNil
	}

	if p.Name == "" {
		return constant.ErrProjectName
	}

	if p.Path == "" {
		return constant.ErrProjectPath
	}

	if p.Source == nil {
		return constant.ErrProjectSource
	}

	if err := p.Source.Validate(); err != nil {
		return constant.ErrProjectSource
	}

	return nil
}

// Convert Project to core Project
func (p *Project) ConvertToCore() (*v1.Project, error) {
	return &v1.Project{
		Name:        p.Name,
		Description: &p.Description,
		Path:        p.Path,
		Labels:      map[string]string{},
	}, nil
}
