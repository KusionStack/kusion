package generator

import (
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
)

// Generator represents a way to generate Intent. Usually, it is implemented by KCL, but we make it as an interface for a more general usage.
// Anyone who implements this interface is regarded as a Generator, and can be integrated by the Kusion workflow.
type Generator interface {
	GenerateSpec(o *Options, project *projectstack.Project, stack *projectstack.Stack) (*models.Intent, error)
}

type Options struct {
	// IsKclPkg represents whether the operation is invoked in a KCL package
	IsKclPkg bool

	// WorkDir represent the filesystem path where the operation is invoked
	WorkDir string

	// Filenames represent all file names included in this operation
	Filenames []string

	// Settings are setting args stored in the setting.yaml
	Settings []string

	// Arguments are args used for a specified Generator. All Generator related args should be passed through this field
	Arguments map[string]string

	// NoStyle represents whether to turn on the spinner output style
	NoStyle bool
}
