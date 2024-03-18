package oci

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

// OCIRepositoryPrefix is the prefix used for OCIRepository URLs.
const OCIRepositoryPrefix = "oci://"

// ParseArtifactRef parses a string representing an OCI repository URL.
// If the string is not a valid representation of an OCI repository URL, ParseArtifactRef returns an error.
func ParseArtifactRef(ociURL string) (name.Reference, error) {
	if !strings.HasPrefix(ociURL, OCIRepositoryPrefix) {
		return nil, fmt.Errorf("URL must be in format 'oci://<domain>/<org>/<repo>'")
	}

	url := strings.TrimPrefix(ociURL, OCIRepositoryPrefix)
	ref, err := name.ParseReference(url)
	if err != nil {
		return nil, fmt.Errorf("'%s' invalid URL format: %w", ociURL, err)
	}

	return ref, nil
}
