package entity

import (
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// Workspace represents the specific configuration workspace
type Workspace struct {
	// ID is the id of the workspace.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the workspace.
	Name string `yaml:"name" json:"name"`
	// DisplayName is the human-readable display name.
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	// Description is a human-readable description of the workspace.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Labels are custom labels associated with the workspace.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the workspace.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// CreationTimestamp is the timestamp of the created for the workspace.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the workspace.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
	// Backend is the corresponding backend for this workspace.
	Backend *Backend `yaml:"backend,omitempty" json:"backend,omitempty"`
}

// Validate checks if the workspace is valid.
// It returns an error if the workspace is not valid.
func (w *Workspace) Validate() error {
	if w == nil {
		return constant.ErrWorkspaceNil
	}

	if w.Name == "" {
		return constant.ErrWorkspaceNameEmpty
	}

	if w.Backend == nil {
		return constant.ErrWorkspaceBackendNil
	}

	if len(w.Owners) == 0 {
		return constant.ErrWorkspaceOwnerNil
	}

	return nil
}
