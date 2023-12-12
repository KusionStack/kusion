package intent

import "encoding/json"

type Type string

const (
	Kubernetes Type = "Kubernetes"
	Terraform  Type = "Terraform"
)

const (
	// ResourceExtensionGVK is the key for resource extension, which is used to
	// store the GVK of the resource.
	ResourceExtensionGVK = "GVK"
	// ResourceExtensionKubeConfig is the key for resource extension, which is used
	// to indicate the path of kubeConfig for Kubernetes type resource.
	ResourceExtensionKubeConfig = "kubeConfig"
)

type Resources []Resource

type Resource struct {
	// ID is the unique key of this resource in the whole State.
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	// providerNamespace:providerName:resourceType:resourceName for Terraform resources
	ID string `json:"id" yaml:"id"`

	// Type represents all Runtimes we supported like Kubernetes and Terraform
	Type Type `json:"type" yaml:"type"`

	// Attributes represents all specified attributes of this resource
	Attributes map[string]interface{} `json:"attributes" yaml:"attributes"`

	// DependsOn contains all resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	// Extensions specifies arbitrary metadata of this resource
	Extensions map[string]interface{} `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}

func (r *Resource) ResourceKey() string {
	return r.ID
}

// DeepCopy return a copy of resource
func (r *Resource) DeepCopy() *Resource {
	var out Resource
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	_ = json.Unmarshal(data, &out)
	return &out
}

func (rs Resources) Index() map[string]*Resource {
	m := make(map[string]*Resource)
	for i := range rs {
		m[rs[i].ResourceKey()] = &rs[i]
	}
	return m
}

// GVKIndex returns a map of GVK to resources, for now, only Kubernetes resources.
func (rs Resources) GVKIndex() map[string][]*Resource {
	m := make(map[string][]*Resource)
	for i := range rs {
		resource := &rs[i]
		if resource.Type != Kubernetes {
			continue
		}
		gvk := resource.Extensions[ResourceExtensionGVK].(string)
		m[gvk] = append(m[gvk], resource)
	}
	return m
}

func (rs Resources) Len() int      { return len(rs) }
func (rs Resources) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs Resources) Less(i, j int) bool {
	switch {
	case rs[i].ID != rs[j].ID:
		return rs[i].ID < rs[j].ID
	default:
		return false
	}
}
