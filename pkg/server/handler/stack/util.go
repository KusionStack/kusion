package stack

import (
	"context"
	"path/filepath"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	buildersapi "kusionstack.io/kusion/pkg/engine/api/builders"
	sourceapi "kusionstack.io/kusion/pkg/engine/api/source"
	"kusionstack.io/kusion/pkg/server/util"
)

func buildOptions(workDir string, kpmParam, dryrun bool) (*buildersapi.Options, *engineapi.APIOptions) {
	// Construct intent options
	intentOptions := &buildersapi.Options{
		IsKclPkg:  kpmParam,
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

// getWorkDirFromSource returns the workdir based on the source
// if the source type is local, it will return the path as an absolute path on the local filesystem
// if the source type is remote (git for example), it will pull the source and return the path to the pulled source
func getWorkDirFromSource(ctx context.Context, stack *entity.Stack, project *entity.Project) (string, string, error) {
	logger := util.GetLogger(ctx)
	logger.Info("Getting workdir from stack source...")
	// TODO: Also copy the local workdir to /tmp directory?
	var err error
	directory := ""
	workDir := stack.Path

	if project.Source != nil && project.Source.SourceProvider != constant.SourceProviderTypeLocal {
		logger.Info("Non-local source provider, locating pulled source directory")
		// pull the latest source code
		directory, err = sourceapi.Pull(ctx, project.Source)
		if err != nil {
			return "", "", err
		}
		logger.Info("config pulled from source successfully", "directory", directory)
		workDir = filepath.Join(directory, stack.Path)
	}
	return directory, workDir, nil
}
