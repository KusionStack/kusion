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
	// ResourceURN is the urn of the resource.
	ResourceURN string `yaml:"resourceURN" json:"resourceURN"`
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
	// Extensions is the extensions of the resource.
	Extensions map[string]interface{} `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	// DependsOn is the depends on of the resource.
	DependsOn []string `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
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

type ResourceInfo struct {
	// ResourceType is the type of the resource.
	ResourceType string `yaml:"resourceType" json:"resourceType"`
	// ResourcePlane is the plane of the resource.
	ResourcePlane string `yaml:"resourcePlane" json:"resourcePlane"`
	// ResourceName is the name of the resource.
	ResourceName string `yaml:"resourceName" json:"resourceName"`
	// IAMResourceID is the id of the resource in IAM.
	IAMResourceID string `yaml:"iamResourceID" json:"iamResourceID"`
	// CloudResourceID is the id of the resource in the cloud.
	CloudResourceID string `yaml:"cloudResourceID" json:"cloudResourceID"`
	// Status is the status of the resource.
	Status string `yaml:"status" json:"status"`
}

type ResourceRelation struct {
	DependentResource  string
	DependencyResource string
}

type ResourceGraph struct {
	Resources map[string]ResourceInfo `yaml:"resources" json:"resources"`
	Relations []ResourceRelation      `yaml:"relations" json:"relations"`
	Workload  string                  `yaml:"workload" json:"workload"`
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

func NewResourceGraph() *ResourceGraph {
	return &ResourceGraph{
		Resources: make(map[string]ResourceInfo),
		Relations: []ResourceRelation{},
		Workload:  "",
	}
}

func (rg *ResourceGraph) SetWorkload(workload string) error {
	rg.Workload = workload
	return nil
}

// AddResourceRelation adds a directed edge from parent to child
func (rg *ResourceGraph) AddResourceRelation(dependentResource, dependencyResource string) {
	rg.Relations = append(rg.Relations, ResourceRelation{
		DependentResource:  dependentResource,
		DependencyResource: dependencyResource,
	})
}

// ConstructResourceGraph constructs the resource graph from the resources.
func (rg *ResourceGraph) ConstructResourceGraph(resources []*Resource) error {
	for _, resource := range resources {
		info := ResourceInfo{
			ResourceType:    resource.ResourceType,
			ResourcePlane:   resource.ResourcePlane,
			ResourceName:    resource.ResourceName,
			IAMResourceID:   resource.IAMResourceID,
			CloudResourceID: resource.CloudResourceID,
			Status:          resource.Status,
		}
		rg.Resources[resource.KusionResourceID] = info
		if resource.Extensions[constant.DefaultWorkloadSig] == true {
			rg.SetWorkload(resource.KusionResourceID)
		}
		if resource.DependsOn != nil {
			for _, dependent := range resource.DependsOn {
				rg.AddResourceRelation(dependent, resource.KusionResourceID)
			}
		}
	}
	return nil
}
