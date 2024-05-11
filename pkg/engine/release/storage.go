package release

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// Storage is used to provide storage service for multiple Releases of a specified Project
// and Workspace.
type Storage interface {
	// Get returns a specified Release by Revision.
	Get(revision uint64) (*v1.Release, error)

	// GetRevisions returns all the Revisions.
	GetRevisions() []uint64

	// GetStackBoundRevisions returns the Revisions of a specified Stack.
	GetStackBoundRevisions(stack string) []uint64

	// GetLatestRevision returns the latest State which corresponds to the current infra Resources.
	GetLatestRevision() uint64

	// Create creates a new Release in the Storage.
	Create(release *v1.Release) error

	// Update updates an existing Release in the Storage.
	Update(release *v1.Release) error
}
