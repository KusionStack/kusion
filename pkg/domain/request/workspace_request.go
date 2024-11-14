package request

import (
	"net/http"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// CreateWorkspaceRequest represents the create request structure for
// workspace.
type CreateWorkspaceRequest struct {
	// Name is the name of the workspace.
	Name string `json:"name" binding:"required"`
	// Description is a human-readable description of the workspace.
	Description string `json:"description"`
	// Labels are custom labels associated with the workspace.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the workspace.
	Owners []string `json:"owners" binding:"required"`
	// BackendID is the configuration backend id associated with the workspace.
	BackendID uint `json:"backendID" binding:"required"`
}

// UpdateWorkspaceRequest represents the update request structure for
// workspace.
type UpdateWorkspaceRequest struct {
	// ID is the id of the workspace.
	ID uint `json:"id" binding:"required"`
	// Name is the name of the workspace.
	Name string `json:"name"`
	// Description is a human-readable description of the workspace.
	Description string `json:"description"`
	// Labels are custom labels associated with the workspace.
	Labels map[string]string `json:"labels"`
	// Owners is a list of owners for the workspace.
	Owners []string `json:"owners" binding:"required"`
	// BackendID is the configuration backend id associated with the workspace.
	BackendID uint `json:"backendID" binding:"required"`
}

type WorkspaceCredentials struct {
	KubeConfigContent string `json:"kubeConfigContent,omitempty"`
	KubeConfigPath    string `json:"kubeConfigPath,omitempty"`
	AliCloudAccessKey string `json:"alicloudAccessKey,omitempty"`
	AliCloudSecretKey string `json:"alicloudSecretKey,omitempty"`
	AliCloudRegion    string `json:"alicloudRegion,omitempty"`
	AwsAccessKey      string `json:"awsAccessKey,omitempty"`
	AwsSecretKey      string `json:"awsSecretKey,omitempty"`
	AwsRegion         string `json:"awsRegion,omitempty"`
}

type WorkspaceConfigs struct {
	*v1.Workspace `yaml:",inline" json:",inline"`
}

func (payload *CreateWorkspaceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *UpdateWorkspaceRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *WorkspaceCredentials) Decode(r *http.Request) error {
	return decode(r, payload)
}

func (payload *WorkspaceConfigs) Decode(r *http.Request) error {
	return decode(r, payload)
}
