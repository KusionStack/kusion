package graph

import (
	"fmt"
	"strings"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/log"
)

const (
	AWSProviderType      = "aws"
	AliCloudProviderType = "alicloud"
	AzureProviderType    = "azure"
	GoogleProviderType   = "google"
	AntProviderType      = "ant"
	// AntProviderRegistrySuffix = "alipay.com"
	CustomProviderType = "custom"
	workloadCategory   = "workload"
	dependencyCategory = "dependency"
	otherCategory      = "other"
)

type ResourceInfo struct {
	ResourceType    string
	CloudResourceID string
	ResourceName    string
}

// addGraphResource adds a GraphResource to the Graph
func addGraphResource(gr *v1.GraphResources, resource *v1.GraphResource, category string) {
	var resourceCollection map[string]*v1.GraphResource
	switch category {
	case workloadCategory:
		if gr.WorkloadResources == nil {
			gr.WorkloadResources = map[string]*v1.GraphResource{}
		}
		resourceCollection = gr.WorkloadResources
	case dependencyCategory:
		if gr.DependencyResources == nil {
			gr.DependencyResources = map[string]*v1.GraphResource{}
		}
		resourceCollection = gr.DependencyResources
	case otherCategory:
		if gr.OtherResources == nil {
			gr.OtherResources = map[string]*v1.GraphResource{}
		}
		resourceCollection = gr.OtherResources
	}

	// Add the resource to the selected collection.
	resourceCollection[resource.ID] = resource
	if gr.ResourceIndex == nil {
		gr.ResourceIndex = map[string]*v1.ResourceEntry{}
	}
	// Update the global resource index with the new resource and its category.
	gr.ResourceIndex[resource.ID] = &v1.ResourceEntry{
		Resource: resource,
		Category: resourceCollection,
	}
}

// FindGraphResourceByID searches for a GraphResource by its ID in the resource index.
// If the resource is found, it returns the corresponding GraphResource. Otherwise, it returns nil.
func FindGraphResourceByID(gr *v1.GraphResources, id string) *v1.GraphResource {
	if entry, found := gr.ResourceIndex[id]; found {
		return entry.Resource
	}

	return nil
}

// FindGraphResourceCollectionByID retrieves the resource collection for a specific resource ID.
// It returns the collection (category) to which the resource belongs or nil if the resource is not found.
func FindGraphResourceCollectionByID(gr *v1.GraphResources, id string) map[string]*v1.GraphResource {
	if gr.ResourceIndex == nil {
		return nil
	}

	if entry, found := gr.ResourceIndex[id]; found {
		return entry.Category
	}

	return nil
}

// GenerateGraph generate a new graph from resources in the spec before apply operation is applied.
func GenerateGraph(resources v1.Resources, gph *v1.Graph) (*v1.Graph, error) {
	log.Infof("Adding spec resources to graph...")
	if gph.Resources == nil {
		gph.Resources = &v1.GraphResources{
			WorkloadResources:   map[string]*v1.GraphResource{},
			DependencyResources: map[string]*v1.GraphResource{},
			OtherResources:      map[string]*v1.GraphResource{},
			ResourceIndex:       map[string]*v1.ResourceEntry{},
		}
	}
	// Get workload resources and its dependsOn resources
	for _, res := range resources {
		// Only add to graph if not exist
		if isResourceWorkload(&res) && FindGraphResourceByID(gph.Resources, res.ID) == nil {
			workload := &v1.GraphResource{
				ID:     res.ID,
				Status: v1.ApplyFail, // status initialized to failure before actual apply operation
			}
			addGraphResource(gph.Resources, workload, workloadCategory)

			if len(res.DependsOn) != 0 {
				for _, dependOn := range res.DependsOn {
					// Only add to graph if not exist
					if FindGraphResourceByID(gph.Resources, dependOn) == nil {
						dependOn := &v1.GraphResource{
							ID:     dependOn,
							Status: v1.ApplyFail, // status initialized to failure before actual apply operation
						}
						addGraphResource(gph.Resources, dependOn, dependencyCategory)
					}
				}
			}
			log.Infof("Added workload resource %s to graph", workload.ID)
			break
		}
	}

	// Put other resources to graph
	for _, res := range resources {
		// Only add to graph if not exist
		if FindGraphResourceByID(gph.Resources, res.ID) == nil {
			other := &v1.GraphResource{
				ID:     res.ID,
				Status: v1.ApplyFail, // status initialized to failure before actual apply operation
			}
			addGraphResource(gph.Resources, other, otherCategory)
			log.Infof("Added workload irrelevant resource %s to graph", other.ID)
		}
	}

	return gph, nil
}

// RemoveResource removes a GraphResource from its category and the global resource index.
func RemoveResource(gph *v1.Graph, resource *v1.GraphResource) {
	// Remove the resource from the category it belongs to
	delete(gph.Resources.ResourceIndex[resource.ID].Category, resource.ID)
	delete(gph.Resources.ResourceIndex, resource.ID)
	resource = nil
}

// RemoveResourceIndex clears the entire resource index of the Graph.
func RemoveResourceIndex(gph *v1.Graph) {
	if gph == nil || gph.Resources == nil {
		return
	}
	gph.Resources.ResourceIndex = nil
}

// isResourceWorkload checks if a resource is identified as a workload.
// It looks for the 'FieldIsWorkload' extension in the resource metadata.
func isResourceWorkload(res *v1.Resource) bool {
	if res.Extensions != nil {
		isWorkload := res.Extensions[v1.FieldIsWorkload]
		if isWorkload != nil && isWorkload.(bool) {
			return true
		}
	}

	return false
}

// GetResourceInfo gets all the essential information for a resource to populate into the resource graph.
func GetResourceInfo(resource *v1.Resource) (*ResourceInfo, error) {
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	// providerNamespace:providerName:resourceType:resourceName for Terraform resources.

	// Meta determines whether this is a Kubernetes resource or Terraform resource.
	resourceTypeMeta := resource.Type
	var resourceType, resourcePlane, cloudResourceID, resourceName string
	// Split the resource name to get the parts
	idParts := strings.Split(resource.ID, ":")
	if len(idParts) != 4 {
		// This indicates a Kubernetes resource without the namespace.
		if len(idParts) == 3 && resource.Type == v1.Kubernetes {
			modifiedID := fmt.Sprintf("%s:%s:%s:%s", idParts[0], idParts[1], "", idParts[2])
			idParts = strings.Split(modifiedID, ":")
		} else {
			return nil, fmt.Errorf("invalid resource ID: %s", resource.ID)
		}
	}

	// Determine resource plane and resource type based on meta type.
	switch resourceTypeMeta {
	case v1.Kubernetes:
		resourcePlane = string(v1.Kubernetes)
		// if this is Kubernetes resource, resource type is apiVersion/kind, resource name is namespace/name.
		resourceType = fmt.Sprintf("%s:%s", idParts[0], idParts[1])
		if idParts[2] == "" {
			resourceName = idParts[3]
		} else {
			resourceName = fmt.Sprintf("%s/%s", idParts[2], idParts[3])
		}
	case v1.Terraform:
		// Get provider info for terraform resources.
		// Look at second element of the id to determine the resource plane.
		switch idParts[1] {
		case AWSProviderType:
			resourcePlane = AWSProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if arn, ok := resource.Attributes["arn"].(string); ok {
				cloudResourceID = arn
			}
		case AzureProviderType:
			resourcePlane = AzureProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if resID, ok := resource.Attributes["id"].(string); ok {
				cloudResourceID = resID
			}
		case GoogleProviderType:
			resourcePlane = GoogleProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if resID, ok := resource.Attributes["id"].(string); ok {
				cloudResourceID = resID
			}
		case AliCloudProviderType:
			resourcePlane = AliCloudProviderType
			resourceType = idParts[2]
			resourceName = idParts[3]
			if resID, ok := resource.Attributes["id"].(string); ok {
				cloudResourceID = resID
			}
		default:
			if _, ok := resource.Extensions["provider"]; ok {
				resourcePlane = CustomProviderType
				resourceType = idParts[2]
			}
		}
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceTypeMeta)
	}

	return &ResourceInfo{
		ResourceType:    fmt.Sprintf("%s:%s", resourcePlane, resourceType),
		CloudResourceID: cloudResourceID,
		ResourceName:    resourceName,
	}, nil
}

// UpdateResourceIndex updates the global resource index for all resources in GraphResources.
func UpdateResourceIndex(graphResources *v1.GraphResources) {
	if graphResources.WorkloadResources != nil {
		for _, resources := range graphResources.WorkloadResources {
			addGraphResource(graphResources, resources, workloadCategory)
		}
	}

	if graphResources.DependencyResources != nil {
		for _, resources := range graphResources.DependencyResources {
			addGraphResource(graphResources, resources, dependencyCategory)
		}
	}

	if graphResources.OtherResources != nil {
		for _, resources := range graphResources.OtherResources {
			addGraphResource(graphResources, resources, otherCategory)
		}
	}
}
