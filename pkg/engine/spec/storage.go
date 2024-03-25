package spec

import v1 "kusionstack.io/kusion/pkg/apis/core/v1"

// Storage is an interface that may be implemented by the application
// to retrieve or create spec entries from storage.
type Storage interface {
	// Get returns the Spec, if the Spec does not exist, return nil.
	Get() (*v1.Intent, error)

	// Apply updates the spec if already exists, or create a new spec.
	Apply(state *v1.Intent) error
}
