package project

// Storage is used to provide storage service for multiple Releases of a specified Project
// and Workspace.
type Storage interface {
	// Get returns a specified Graph.
	Get() (map[string][]string, error)
}
