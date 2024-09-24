package graph

import v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"

// Storage is used to provide storage service for multiple Releases of a specified Project
// and Workspace.
type Storage interface {
	// Get returns a specified Graph.
	Get() (*v1.Graph, error)

	// Create creates a new Graph in the Storage.
	Create(*v1.Graph) error

	// Update updates an existing Graph in the Storage.
	Update(*v1.Graph) error

	// Delete deletes an existing Graph in the Storage.
	Delete() error

	// CheckGraphStorageExistence checks if the Graph storage exists.
	CheckGraphStorageExistence() bool
}
