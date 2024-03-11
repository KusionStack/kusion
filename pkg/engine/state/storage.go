package state

import (
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// Storage is used to provide the state storage for a set of real resources belonging to a specified stack,
// which is determined by the combination of project name, stack name and workspace name.
type Storage interface {
	// Get returns the state, if the state does not exist, return nil.
	Get() (*v1.State, error)

	// Apply updates the state if already exists, or create a new state.
	Apply(state *v1.State) error
}
