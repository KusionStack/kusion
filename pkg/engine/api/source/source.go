package source

import (
	"context"
	"os"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	sp "kusionstack.io/kusion/pkg/domain/entity/source_providers"
	"kusionstack.io/kusion/pkg/server/util"
)

// Pull() is a method that pulls the source code from the git source provider.
func Pull(ctx context.Context, source *entity.Source) (string, error) {
	if source == nil {
		return "", constant.ErrSourceNil
	}

	// Create a new GitSourceProvider with the remote URL and pulls into /tmp directory.
	localDirectory, err := os.MkdirTemp("/tmp", "kcp-kusion-")
	if err != nil {
		return "", err
	}
	gsp := sp.NewGitSourceProvider(source.Remote.String(), localDirectory, "")

	// Call the Get() method of the source provider to pull the source code.
	directory, err := gsp.Get(ctx, entity.WithType(constant.SourceProviderTypeGit))
	if err != nil {
		return "", err
	}
	return directory, nil
}

// Cleanup() is a method that cleans up the temporary source code from the source provider.
func Cleanup(ctx context.Context, localDirectory string) {
	logger := util.GetLogger(ctx)
	logger.Info("Cleaning up temp directory...")
	if localDirectory == "" {
		return
	}
	gsp := sp.NewGitSourceProvider("", localDirectory, "")
	// Call the Cleanup() method of the source provider to clean up the source code.
	gsp.Cleanup(ctx)
}
