package entity

import (
	"fmt"
	"time"

	"kusionstack.io/kusion/pkg/domain/constant"
)

// Resource represents the specific configuration code resource,
// which should be a specific instance of the resource provider.
type Resource struct {
	// ID is the id of the resource.
	ID uint `yaml:"id" json:"id"`
	// Stack is the stack associated with the resource.
	Stack *Stack `yaml:"stack" json:"stack"`
	// ResourceType is the type of the resource.
	ResourceType string `yaml:"resourceType" json:"resourceType"`
	// ResourcePlane is the plane of the resource.
	ResourcePlane string `yaml:"resourcePlane" json:"resourcePlane"`
	// ResourceName is the name of the resource.
	ResourceName string `yaml:"resourceName" json:"resourceName"`
	// KusionResourceID is the id of the resource in Kusion.
	KusionResourceID string `yaml:"kusionResourceID" json:"kusionResourceID"`
	// IAMResourceID is the id of the resource in IAM.
	IAMResourceID string `yaml:"iamResourceID" json:"iamResourceID"`
	// CloudResourceID is the id of the resource in the cloud.
	CloudResourceID string `yaml:"cloudResourceID" json:"cloudResourceID"`
	// LastAppliedRevision is the revision of the last sync.
	LastAppliedRevision string `yaml:"LastAppliedRevision" json:"LastAppliedRevision"`
	// LastAppliedTimestamp is the timestamp of the last sync.
	LastAppliedTimestamp time.Time `yaml:"LastAppliedTimestamp" json:"LastAppliedTimestamp"`
	// Status is the status of the resource.
	Status string `yaml:"status" json:"status"`
	// Attributes is the attributes of the resource.
	Attributes map[string]interface{} `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	// Provider is the provider of the resource.
	Provider string `yaml:"provider" json:"provider"`
	// Labels are custom labels associated with the resource.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the resource.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// CreationTimestamp is the timestamp of the created for the resource.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the resource.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
}

type ResourceFilter struct {
	OrgID            uint
	ProjectID        uint
	StackID          uint
	ResourcePlane    string
	ResourceType     string
	KusionResourceID string
	Status           string
}

// Validate checks if the resource is valid.
// It returns an error if the resource is not valid.
func (r *Resource) Validate() error {
	if r == nil {
		return fmt.Errorf("resource is nil")
	}

	if r.Stack == nil {
		return constant.ErrResourceHasNilStack
	}

	return nil
}
