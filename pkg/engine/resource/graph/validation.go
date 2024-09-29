package graph

import (
	"errors"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

var (
	ErrEmptyGraph         = errors.New("empty graph")
	ErrEmptyProject       = errors.New("empty project")
	ErrEmptyWorkspace     = errors.New("empty workspace")
	ErrEmptyResources     = errors.New("empty resources")
	ErrEmptyResourceIndex = errors.New("empty resource index")
)

// ValidateGraph checks the validity of a Graph object.
// It ensures that the essential fields within the Graph structure are not empty or nil.
// If any of the required fields are missing, an appropriate error is returned.
func ValidateGraph(graph *v1.Graph) error {
	if graph == nil {
		return ErrEmptyGraph
	}
	if graph.Project == "" {
		return ErrEmptyProject
	}
	if graph.Workspace == "" {
		return ErrEmptyWorkspace
	}
	if graph.Resources == nil {
		return ErrEmptyResources
	}

	if graph.Resources.ResourceIndex == nil {
		return ErrEmptyResourceIndex
	}
	return nil
}
