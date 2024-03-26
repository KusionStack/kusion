package request

import (
	"kusionstack.io/kusion/pkg/domain/service"
)

// CreateStackRequest represents the create request structure for
// stack.
type CreateStackRequest struct {
	// Name is the name of the stack.
	Name string `json:"name" binding:"required"`
	// SourceID is the configuration source id associated with the stack.
	// SourceID uint `json:"sourceID,string" binding:"required"`
	// ProjectID is the project id of the stack within the source.
	ProjectID uint `json:"projectID,string" binding:"required"`
	// OrganizationID is the organization id associated with the stack.
	// OrganizationID uint `json:"organizationID,string" binding:"required"`
	// Path is the relative path of the stack within the source.
	Path string `json:"path" binding:"required"`
	// DesiredVersion is the desired revision of stack.
	DesiredVersion string `json:"desiredVersion" binding:"required"`
	// Description is a human-readable description of the stack.
	Description string `json:"description"`
	// Labels are custom labels associated with the stack.
	Labels []string `json:"labels"`
	// Owners is a list of owners for the stack.
	Owners []string `json:"owners"`
}

// UpdateStackRequest represents the update request structure for
// stack.
type UpdateStackRequest struct {
	// ID is the id of the stack.
	ID uint `json:"id" binding:"required"`
	// Name is the name of the stack.
	Name string `json:"name"`
	// SourceID is the configuration source id associated with the stack.
	// SourceID uint `json:"sourceID,string"`
	// ProjectID is the project id of the stack within the stack.
	ProjectID uint `json:"projectID,string"`
	// OrganizationID is the organization id associated with the stack.
	// OrganizationID uint `json:"organizationID,string"`
	// Path is the relative path of the stack within the stack.
	Path string `json:"path"`
	// DesiredVersion is the desired revision of stack.
	DesiredVersion string `json:"desiredVersion"`
	// Description is a human-readable description of the stack.
	Description string `json:"description"`
	// Labels are custom labels associated with the stack.
	Labels map[string]string `json:"labels"`
	// Owners is a list of owners for the stack.
	Owners []string `json:"owners"`
}

// ExecuteStackRequest is the common request for preview and sync operation.
type ExecuteStackRequest struct {
	// SourceProviderType is the type of the source provider.
	SourceProviderType string `json:"sourceProviderType"`
	// Remote is the remote url of the stack to be pulled.
	Remote string `json:"remote" binding:"required"`
	// Version is the version of the stack to be pulled.
	Version string `json:"version" binding:"required"`
	// Envs lets you set the env when executes in the form "key=value".
	Envs []string `json:"envs,omitempty" yaml:"envs,omitempty"`
	// AdditionalPaths is the additional paths to be added to the stack.
	AdditionalPaths []string `json:"additionalPaths,omitempty"`
	// DisableState is the flag to disable state management.
	DisableState bool `json:"disableState,omitempty"`
	// Extensions is the extensions for the stack request.
	Extensions service.Extensions `json:"extensions,omitempty"`
}

type ExecutePreviewStackRequest struct {
	// OutputFormat specify the output format, one of "", "json".
	OutputFormat string `json:"outputFormat,omitempty" binding:"oneof='' 'json'"`
	// DriftMode is a boolean field used to represent the state of the drift mode.
	DriftMode bool `json:"driftMode,omitempty" yaml:"driftMode,omitempty"`
}

// PreviewStackRequest represents the preview request structure
// for stack.
type PreviewStackRequest struct {
	ExecuteStackRequest        `json:",inline"`
	ExecutePreviewStackRequest `json:",inline"`
	StackPath                  string `json:"stackPath" binding:"required"`
}

// SyncStackRequest represents the sync request structure
// for stack.
type SyncStackRequest struct {
	ExecuteStackRequest `json:",inline"`
	StackPath           string `json:"stackPath" binding:"required"`
}

// PreviewStacksRequest represents the preview request structure
// for stacks.
type PreviewStacksRequest struct {
	ExecuteStackRequest        `json:",inline"`
	ExecutePreviewStackRequest `json:",inline"`
	StackPaths                 []string `json:"stackPaths" binding:"required"`
}

// SyncStacksRequest represents the sync request structure
// for stacks.
type SyncStacksRequest struct {
	ExecuteStackRequest `json:",inline"`
	StackPaths          []string `json:"stackPaths" binding:"required"`
}

// InspectStacksRequest represents the inspect request structure
// for stacks.
type InspectStacksRequest struct {
	Verbose    bool     `json:"verbose,omitempty" yaml:"verbose,omitempty"`
	Remote     string   `json:"remote" yaml:"remote" binding:"required"`
	StackPaths []string `json:"stackPaths" yaml:"stackPaths" binding:"required"`
}
