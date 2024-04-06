package workspace

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// Storage is used to provide the storage service for multiple workspaces.
type Storage interface {
	// Get returns the workspace configurations. If name is not specified, get the current workspace
	// configurations.
	Get(name string) (*v1.Workspace, error)

	// Create the workspace.
	Create(ws *v1.Workspace) error

	// Update the workspace. If name is not specified, updates the current workspace, and set the current
	// workspace name in the input's name field.
	Update(ws *v1.Workspace) error

	// Delete deletes the workspace. If name is not specified, deletes the current workspace.
	Delete(name string) error

	// GetNames returns the names of all the existing workspaces.
	GetNames() ([]string, error)

	// GetCurrent gets the name of the current workspace.
	GetCurrent() (string, error)

	// SetCurrent sets the specified workspace as the current workspace.
	SetCurrent(name string) error
}
