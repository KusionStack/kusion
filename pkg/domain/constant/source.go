package constant

import (
	"errors"
	"fmt"
)

// SourceProviderType represents the type of varying source providers,
// source provider is the general abstraction of version control systems (VCS),
// also known as source control systems (SCM).
type SourceProviderType string

var (
	ErrSourceNil               = errors.New("source is nil")
	ErrDirectoryToCleanupEmpty = errors.New("temp kcp-kusion directory to clean up is empty")
)

const (
	// SourceProviderTypeGit represents git source provider type.
	SourceProviderTypeGit SourceProviderType = "git"

	// SourceProviderTypeGithub represents github source provider type.
	SourceProviderTypeGithub SourceProviderType = "github"

	// SourceProviderTypeOCI represents oci source provider type.
	SourceProviderTypeOCI SourceProviderType = "oci"

	// SourceProviderTypeLocal represents local source provider type.
	SourceProviderTypeLocal SourceProviderType = "local"
)

// ParseSourceProviderType parses a string into a SourceProviderType.
// If the string is not a valid SourceProviderType, it returns an error.
func ParseSourceProviderType(s string) (SourceProviderType, error) {
	switch s {
	case string(SourceProviderTypeGit):
		return SourceProviderTypeGit, nil
	case string(SourceProviderTypeGithub):
		return SourceProviderTypeGithub, nil
	case string(SourceProviderTypeOCI):
		return SourceProviderTypeOCI, nil
	case string(SourceProviderTypeLocal):
		return SourceProviderTypeLocal, nil
	default:
		return SourceProviderType(""), fmt.Errorf("invalid SourceProviderType: %q", s)
	}
}
