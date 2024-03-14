package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"

	meta "kusionstack.io/kusion/pkg/oci/metadata"
)

// Push takes care of the actual artifact push behavior. It performs following operations:
// - builds tarball from given directory also corresponding layer
// - adds this layer to an empty OpenContainers artifact
// - annotates the artifact with the given annotations
// - uploads the final artifact to the OCI registry
// - returns the digest URL of the upstream artifact
func (c *Client) Push(ctx context.Context, registryURL, sourceDir string, metadata meta.Metadata, ignorePaths []string) (string, error) {
	ref, err := parseArtifactRef(registryURL)
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "oci")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "artifact.tgz")
	if err := c.Build(tmpFile, sourceDir, ignorePaths); err != nil {
		return "", err
	}

	// Add missing metadata
	if metadata.Created == "" {
		ct := time.Now().UTC()
		metadata.Created = ct.Format(time.RFC3339)
	}

	image := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	image = mutate.ConfigMediaType(image, CanonicalConfigMediaType)
	image = mutate.Annotations(image, metadata.ToAnnotations()).(v1.Image)

	layer, err := tarball.LayerFromFile(tmpFile, tarball.WithMediaType(CanonicalContentMediaType))
	if err != nil {
		return "", fmt.Errorf("creating content layer failed: %w", err)
	}

	image, err = mutate.Append(image, mutate.Addendum{Layer: layer})
	if err != nil {
		return "", fmt.Errorf("appeding content to artifact failed: %w", err)
	}

	if err := crane.Push(image, registryURL, c.optionsWithContext(ctx)...); err != nil {
		return "", fmt.Errorf("pushing artifact failed: %w", err)
	}

	digest, err := image.Digest()
	if err != nil {
		return "", fmt.Errorf("parsing artifact digest failed: %w", err)
	}

	return ref.Context().Digest(digest.String()).String(), err
}
