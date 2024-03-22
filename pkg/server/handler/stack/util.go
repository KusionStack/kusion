package stack

import (
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	buildersapi "kusionstack.io/kusion/pkg/engine/api/builders"
)

func buildOptions(workDir string, isKCLPackageParam, dryrun bool) (*buildersapi.Options, *engineapi.APIOptions) {
	// Construct intent options
	intentOptions := &buildersapi.Options{
		IsKclPkg:  isKCLPackageParam,
		WorkDir:   workDir,
		Arguments: map[string]string{},
		NoStyle:   true,
	}
	// Construct preview api option
	// TODO: Complete preview options
	// TODO: Operator should be derived from auth info
	// TODO: Cluster should be derived from workspace config
	previewOptions := &engineapi.APIOptions{
		// Operator:     "operator",
		// Cluster:      "cluster",
		// IgnoreFields: []string{},
		DryRun: dryrun,
	}
	return intentOptions, previewOptions
}
