package release

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// Storage is used to provide storage service for multiple releases.
type Storage interface {
	// Get returns a specified Release which is determined by the group of Project, Workspace
	// and Revision.
	Get(project, workspace string, revision uint64) (*v1.Release, error)

	// GetRevisions returns the Revisions of a specified Project and Workspace.
	GetRevisions(project, workspace string) ([]uint64, error)

	// GetStackBoundRevisions returns the Revisions of a specified Project, Stack and Workspace.
	GetStackBoundRevisions(project, stack, workspace string) ([]uint64, error)

	// GetLatestRevision returns the latest State which corresponds to the current infra Resources.
	GetLatestRevision(project, workspace string) (uint64, error)

	// Create creates a new Release in the Storage.
	Create(release *v1.Release) error

	// Update updates an existing Release in the Storage.
	Update(release *v1.Release) error
}
