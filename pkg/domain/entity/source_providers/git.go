package sourceproviders

// This file should contain the git implementation for the sourceProvider interface

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"

	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/server/util"
)

var _ entity.SourceProvider = &GitSourceProvider{}

// GitSourceProvider is the implementation of the SourceProvider interface
type GitSourceProvider struct {
	// The remote URL of the git repository
	Remote string
	// The directory to clone the git repository
	Directory string
	// The version of the git repository
	Version string
}

// NewGitSourceProvider creates a new GitSourceProvider
func NewGitSourceProvider(remote, directory, version string) *GitSourceProvider {
	return &GitSourceProvider{
		Remote:    remote,
		Directory: directory,
		Version:   version,
	}
}

// Type returns the type of the source provider
func (g *GitSourceProvider) Type() constant.SourceProviderType {
	return constant.SourceProviderTypeGit
}

// Get clones the git repository and returns the directory
func (g *GitSourceProvider) Get(ctx context.Context, opts ...entity.GetOption) (string, error) {
	// Create the directory if it does not exist
	if _, err := os.Stat(g.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(g.Directory, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// TODO: Add basic and ssh authentication
	// token := os.Getenv("GITHUB_TOKEN")
	// Use golang git library to clone the repository
	repo, err := git.PlainCloneContext(ctx, g.Directory, false, &git.CloneOptions{
		URL:      g.Remote,
		Progress: os.Stdout,
		// Auth: &http.TokenAuth{
		// 	Token: token,
		// },
	})
	if err != nil {
		return "", err
	}
	log.Info("Successfully cloned git repository: %s", repo)

	if g.Version != "" {
		err := checkoutRevision(g.Directory, g.Version)
		if err != nil {
			return "", ErrCheckingOutRevision
		}
	}

	// // Clone the git repository
	// cmd := exec.CommandContext(ctx, "git", "clone", g.Remote, g.Directory)
	// if err := cmd.Run(); err != nil {
	// 	return "", fmt.Errorf("failed to clone git repository: %w", err)
	// }

	// Checkout the version
	// if g.Version != "" {
	// 	cmd = exec.CommandContext(ctx, "git", "-C", g.Directory, "checkout", g.Version)
	// 	if err := cmd.Run(); err != nil {
	// 		return "", fmt.Errorf("failed to checkout version: %w", err)
	// 	}
	// }

	log.Infof("Successfully cloned git repository: %s", g.Remote)

	return g.Directory, nil
}

// Cleanup cleans up the resources of the provider
func (g *GitSourceProvider) Cleanup(ctx context.Context) {
	logger := util.GetLogger(ctx)
	logger.Info("Cleaning up temp kcp-kusion directory...")

	// Remove the directory
	if err := os.RemoveAll(g.Directory); err != nil {
		log.Errorf("failed to remove directory: %v", err)
	}
	logger.Info("temp directory removed", "directory", g.Directory)
}

// GetGitSourceProvider returns a GitSourceProvider
func GetGitSourceProvider(ctx context.Context, remote, directory, version string) (*GitSourceProvider, error) {
	// Get the absolute path of the directory
	absPath, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return NewGitSourceProvider(remote, absPath, version), nil
}
