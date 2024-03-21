package constant

import "fmt"

// SourceProviderType represents the type of varying source providers,
// source provider is the general abstraction of version control systems (VCS),
// also known as source control systems (SCM).
type SourceProviderType string

const (
	// SourceProviderTypeGithub represents github source provider type.
	SourceProviderTypeGit    SourceProviderType = "git"
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
