package request

import (
	"net/http"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// CreateProjectRequest represents the create request structure for
// project.
type CreateProjectRequest struct {
	// Name is the name of the project.
	Name string `json:"name"`
	// SourceID is the configuration source id associated with the project.
	SourceID uint `json:"sourceID"`
	// OrganizationID is the organization id associated with the project.
	OrganizationID uint `json:"organizationID"`
	// Description is a human-readable description of the project.
	Description string `json:"description"`
	// Path is the relative path of the project within the sources.
	Path string `json:"path"`
	// Domain is the domain of the project, typically serving as the parent folder name for the project.
	Domain string `json:"domain" binding:"required"`
	// Labels are custom labels associated with the project.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the project.
	Owners []string `json:"owners"`
}

// UpdateProjectRequest represents the update request structure for
// project.
type UpdateProjectRequest struct {
	// ID is the id of the project.
	ID                   uint `json:"id" binding:"required"`
	CreateProjectRequest `json:",inline" yaml:",inline"`
}

func (payload *CreateProjectRequest) Validate() error {
	// Validate domain or organization id is required
	if payload.Domain == "" && payload.OrganizationID == 0 {
		return constant.ErrOrgIDOrDomainRequired
	}

	// Validate project, stack and appconfig name contains only alphanumeric
	// and underscore characters
	if validName(payload.Name) {
		return constant.ErrInvalidProjectName
	}

	// Validate project name should only contain one or more capturing group
	// that contains a backslash with alphanumeric and underscore characters
	if validPath(payload.Path) {
		return constant.ErrInvalidProjectPath
	}
	return nil
}

func (payload *UpdateProjectRequest) Validate() error {
	return payload.CreateProjectRequest.Validate()
}

func (payload *CreateProjectRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateProjectRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
