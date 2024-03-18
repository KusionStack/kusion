package client

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"

	"kusionstack.io/kusion/pkg/oci"
)

// Tag creates a new tag for the given artifact using the same OCI repository as the origin.
func (c *Client) Tag(ctx context.Context, registryURL, tag string) error {
	ref, err := oci.ParseArtifactRef(registryURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return crane.Tag(ref.String(), tag, c.optionsWithContext(ctx)...)
}
