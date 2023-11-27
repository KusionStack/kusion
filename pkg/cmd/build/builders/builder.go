package builders

import (
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
)

// Builder represents a method to build an Intent. Typically, it is implemented by the AppConfigureBuilder,
// but we have designed it as an interface to allow for more general usage. Any struct that implements this interface
// is considered a Builder and can be integrated into the Kusion workflow.
type Builder interface {
	Build(o *Options, project *project.Project, stack *stack.Stack) (*intent.Intent, error)
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

	// Arguments are args used for a specified Builder. All Builder related args should be passed through this field
	Arguments map[string]string

	// NoStyle represents whether to turn on the spinner output style
	NoStyle bool
}
