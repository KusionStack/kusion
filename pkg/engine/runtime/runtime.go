package runtime

import (
	"context"
	"kusionstack.io/kusion/pkg/engine/models"

	"kusionstack.io/kusion/pkg/status"
)

type Runtime interface {
	// Apply resource with planState. priorState is given to Runtime for three-way-merge if it needs
	Apply(ctx context.Context, priorState, planState *models.Resource) (*models.Resource, status.Status)

	// Read the latest state of this resource
	Read(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status)

	// Delete resource
	Delete(ctx context.Context, resourceState *models.Resource) status.Status

	// Watch the latest state or event of this resource. This is very helpful for us to know what is happening when apply resources
	Watch(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status)
}
