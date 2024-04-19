package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
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
// - check if the target reference exists and if it is an image index. Creates a new image index if it does not exist
// - uploads the final artifact to the OCI registry and appends the artifact to the index,
// - returns the digest URL of the upstream artifact
func (c *Client) Push(
	ctx context.Context,
	ociURL, version, sourceDir string,
	metadata meta.Metadata,
	ignorePaths []string,
) (string, string, error) {
	idxURL := fmt.Sprintf("%s:%s", ociURL, version)
	ref, err := oci.ParseArtifactRef(idxURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid OCI repository url: %w", err)
	}

	// Check if the target reference exists and if it is an image index
	refStr := ref.String()
	exists := true
	var base v1.ImageIndex

	manifest, err := crane.Get(refStr, c.opts.craneOptions...)
	if err != nil {
		var t *transport.Error
		ok := errors.As(err, &t)
		if ok && t.StatusCode == 404 {
			exists = false
			base = empty.Index
		} else {
			return "", "", fmt.Errorf("get manifest failed: %s, %w", refStr, err)
		}
	}

	if exists {
		if !manifest.MediaType.IsIndex() {
			return "", "", fmt.Errorf("expected %s to be an index, got %q", refStr, manifest.MediaType)
		}
		base, err = manifest.ImageIndex()
		if err != nil {
			return "", "", fmt.Errorf("get manifest image index failed: %s, %w", refStr, err)
		}
	}

	// build image
	tmpDir, err := os.MkdirTemp("", "oci")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, ArtifactTarballFileName)
	if err = c.Build(tmpFile, sourceDir, ignorePaths); err != nil {
		return "", "", err
	}

	// Add missing metadata
	if metadata.Created == "" {
		ct := time.Now().UTC()
		metadata.Created = ct.Format(time.RFC3339)
	}

	image := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	image = mutate.ConfigMediaType(image, CanonicalConfigMediaType)
	image = mutate.Annotations(image, metadata.ToAnnotations()).(v1.Image)

	platform := metadata.Platform
	if platform == nil {
		return "", "", fmt.Errorf("platform is not set")
	}
	image, err = mutate.ConfigFile(image, &v1.ConfigFile{
		Architecture: platform.Architecture,
		OS:           platform.OS,
	})
	if err != nil {
		return "", "", fmt.Errorf("setting image config file failed: %w", err)
	}

	layer, err := tarball.LayerFromFile(tmpFile, tarball.WithMediaType(CanonicalContentMediaType))
	if err != nil {
		return "", "", fmt.Errorf("creating content layer failed: %w", err)
	}

	image, err = mutate.Append(image, mutate.Addendum{
		Layer: layer,
		Annotations: map[string]string{
			meta.AnnotationTitle: ArtifactTarballFileName,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("appeding content to artifact failed: %w", err)
	}

	imgURL := fmt.Sprintf("%s-%s_%s:%s", ociURL, platform.OS, platform.Architecture, version)
	imgRef, err := oci.ParseArtifactRef(imgURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid image repository url: %w", err)
	}
	if err = crane.Push(image, imgRef.String(), c.optionsWithContext(ctx)...); err != nil {
		return "", "", fmt.Errorf("pushing artifact failed: %w", err)
	}
	imgDigest, err := image.Digest()
	if err != nil {
		return "", "", fmt.Errorf("parsing image digest failed: %w", err)
	}

	cf, err := image.ConfigFile()
	if err != nil {
		return "", "", fmt.Errorf("parsing image config file failed: %w", err)
	}

	newDesc, err := partial.Descriptor(image)
	if err != nil {
		return "", "", fmt.Errorf("parsing image descriptor file failed: %w", err)
	}
	newDesc.Platform = cf.Platform()
	addendum := mutate.IndexAddendum{
		Add:        image,
		Descriptor: *newDesc,
	}
	idx := mutate.AppendManifests(base, addendum)
	idxDigest, err := idx.Digest()
	if err != nil {
		return "", "", fmt.Errorf("parsing index digest failed: %w", err)
	}

	o := crane.GetOptions(c.opts.craneOptions...)
	if err = remote.WriteIndex(ref, idx, o.Remote...); err != nil {
		return "", "", fmt.Errorf("pushing image index %s: %w", refStr, err)
	}

	idxDigestURL := ref.Context().Digest(idxDigest.String()).String()
	imgDigestURL := ref.Context().Digest(imgDigest.String()).String()
	idxURL = fmt.Sprintf("%s%s", oci.OCIRepositoryPrefix, idxDigestURL)
	imgURL = fmt.Sprintf("%s%s", oci.OCIRepositoryPrefix, imgDigestURL)
	return idxURL, imgURL, nil
}
