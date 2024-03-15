package workspace

import (
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// Storage is used to provide the storage service for multiple workspaces.
type Storage interface {
	// Get returns the specified workspace configurations.
	Get(name string) (*v1.Workspace, error)

	// Create the workspace.
	Create(ws *v1.Workspace) error

	// Update the workspace.
	Update(ws *v1.Workspace) error

	// Delete deletes the specified workspace.
	Delete(name string) error

	// Exist returns the specified workspace exists or not.
	Exist(name string) (bool, error)

	// GetNames returns the names of all the existing workspaces.
	GetNames() ([]string, error)

	// GetCurrent gets the name of the current workspace.
	GetCurrent() (string, error)

	// SetCurrent sets the specified workspace as the current workspace.
	SetCurrent(name string) error
}
