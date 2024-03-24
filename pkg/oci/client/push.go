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

	"kusionstack.io/kusion/pkg/oci"
	meta "kusionstack.io/kusion/pkg/oci/metadata"
)

// ArtifactTarballFileName defines name for the generated artifact tarball file
const ArtifactTarballFileName = "artifact.tgz"

// Push takes care of the actual artifact push behavior. It performs following operations:
// - builds tarball from given directory also corresponding layer
// - adds this layer to an empty OpenContainers artifact
// - annotates the artifact with the given annotations
// - uploads the final artifact to the OCI registry
// - returns the digest URL of the upstream artifact
func (c *Client) Push(ctx context.Context, ociURL, sourceDir string, metadata meta.Metadata, ignorePaths []string) (string, error) {
	ref, err := oci.ParseArtifactRef(ociURL)
	if err != nil {
		return "", fmt.Errorf("invalid OCI repository url: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "oci")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, ArtifactTarballFileName)
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

	image, err = mutate.Append(image, mutate.Addendum{
		Layer: layer,
		Annotations: map[string]string{
			meta.AnnotationTitle: ArtifactTarballFileName,
		},
	})
	if err != nil {
		return "", fmt.Errorf("appeding content to artifact failed: %w", err)
	}

	if err := crane.Push(image, ref.String(), c.optionsWithContext(ctx)...); err != nil {
		return "", fmt.Errorf("pushing artifact failed: %w", err)
	}

	digest, err := image.Digest()
	if err != nil {
		return "", fmt.Errorf("parsing artifact digest failed: %w", err)
	}

	digestURL := ref.Context().Digest(digest.String()).String()
	return fmt.Sprintf("%s%s", oci.OCIRepositoryPrefix, digestURL), nil
}
