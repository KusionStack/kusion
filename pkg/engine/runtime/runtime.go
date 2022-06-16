package runtime

import (
	"context"

	"kusionstack.io/kusion/pkg/engine/models"

	"kusionstack.io/kusion/pkg/status"
)

// Runtime represents an actual infrastructure runtime managed by Kusion and every runtime implements this interface can be orchestrated
// by Kusion like normal K8s resources. All methods in this interface are designed for manipulating one resource at a time and will be
// invoked in operations like Apply, Preview, Destroy, etc.
type Runtime interface {
	// Apply means modify this resource to the desired state described in the request,
	// and it will turn into creating or updating a resource in most scenarios.
	// If the infrastructure runtime already provides an Apply method that conform to this method's semantics meaning,
	// like the Kubernetes Runtime, you can directly invoke this method without any conversion.
	// PlanState and priorState are given in this method for the runtime which would make a
	// three-way-merge (planState,priorState and live state) when implementing this interface
	Apply(ctx context.Context, priorState, planState *models.Resource) (*models.Resource, status.Status)

	// Read the latest state of this resource
	Read(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status)

	// Delete this resource in the actual infrastructure and return success if this resource is not exist
	Delete(ctx context.Context, resourceState *models.Resource) status.Status

	// Watch the latest state or event of this resource.
	// This is an optional method for the Runtime to implement,
	// but it will be very helpful for us to know what is happening when applying this resource
	Watch(ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status)
}
