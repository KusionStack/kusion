package request

import (
	"net/http"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
)

// CreateStackRequest represents the create request structure for
// stack.
type CreateStackRequest struct {
	// Name is the name of the stack.
	Name string `json:"name" binding:"required"`
	// ProjectID is the project id of the stack within the source.
	ProjectID uint `json:"projectID"`
	// ProjectName is the project name of the stack within the source.
	ProjectName string `json:"projectName"`
	// Type is the type of the stack.
	Type string `json:"type"`
	// Path is the relative path of the stack within the source.
	Path string `json:"path"`
	// DesiredVersion is the desired revision of stack.
	DesiredVersion string `json:"desiredVersion"`
	// Description is a human-readable description of the stack.
	Description string `json:"description"`
	// Labels are custom labels associated with the stack.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the stack.
	Owners []string `json:"owners"`
}

type UpdateVariableRequest struct {
	// Project is the project related to stack
	Project string `json:"project,omitempty"`
	// Path is the relative path of the stack within the source.
	Path     string `json:"path,omitempty"`
	IsSecret bool   `json:"isSecret,omitempty"`
	// key is the unique index to use value in specific stack
	Key string `json:"key,omitempty"`
	// value is the plain value of no sensitive data
	Value       string              `json:"value,omitempty"`
	SecretValue *entity.SecretValue `json:"secretValue,omitempty"`
}

// UpdateStackRequest represents the update request structure for
// stack.
type UpdateStackRequest struct {
	// ID is the id of the stack.
	ID uint `json:"id" binding:"required"`
	// Name is the name of the stack.
	Name string `json:"name"`
	// ProjectID is the project id of the stack within the source.
	ProjectID uint `json:"projectID"`
	// ProjectName is the project name of the stack within the source.
	ProjectName string `json:"projectName"`
	// Type is the type of the stack.
	Type string `json:"type"`
	// Path is the relative path of the stack within the source.
	Path string `json:"path"`
	// DesiredVersion is the desired revision of stack.
	DesiredVersion string `json:"desiredVersion"`
	// Description is a human-readable description of the stack.
	Description string `json:"description"`
	// Labels are custom labels associated with the stack.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the stack.
	Owners []string `json:"owners"`
}

func (payload *CreateStackRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateStackRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateVariableRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *CreateStackRequest) Validate() error {
	if payload.ProjectID == 0 && payload.ProjectName == "" {
		return constant.ErrProjectNameOrIDRequired
	}

	// Validate project, stack and appconfig name contains only alphanumeric
	// and underscore characters
	if validName(payload.Name) {
		return constant.ErrInvalidStackName
	}
	if payload.ProjectName != "" && validName(payload.ProjectName) {
		return constant.ErrInvalidProjectName
	}

	// Validate stack path if provided,
	// It will be set in the stack manager if not provided
	if payload.Path != "" && validPath(payload.Path) {
		return constant.ErrInvalidProjectPath
	}

	return nil
}

func (payload *UpdateStackRequest) Validate() error {
	// Validate project, stack and appconfig name contains only alphanumeric
	// and underscore characters
	if payload.Name != "" && validName(payload.Name) {
		return constant.ErrInvalidStackName
	}

	if payload.ProjectName != "" && validName(payload.ProjectName) {
		return constant.ErrInvalidProjectName
	}

	// Validate path should only contain one or more capturing group
	// that contains a backslash with alphanumeric and underscore characters
	if payload.Path != "" && validPath(payload.Path) {
		return constant.ErrInvalidProjectPath
	}

	return nil
}
