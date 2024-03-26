package builders

import (
	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// Builder represents a method to build an Intent. Typically, it is implemented by the AppConfigureBuilder,
// but we have designed it as an interface to allow for more general usage. Any struct that implements this interface
// is considered a Builder and can be integrated into the Kusion workflow.
type Builder interface {
	Build(kclPackage *api.KclPackage, project *v1.Project, stack *v1.Stack) (*v1.Intent, error)
}

type Options struct {
	// KclPkg represents the kcl package information. If it is nil, it means this workdir is not a kcl package
	KclPkg *api.KclPackage

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
