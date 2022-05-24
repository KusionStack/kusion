package models

type Type string

type Resources []Resource

type Resource struct {
	// ID is the unique key of this resource in the whole State. ApiVersion:Kind:Namespace:Name is an idiomatic way of Kubernetes resources.
	ID string `json:"id"`

	// Type represents all Runtimes we supported
	Type Type `json:"type"`

	//Attributes represents all specified attributes of this resource
	Attributes map[string]interface{} `json:"attributes"`

	// DependsOn contains all resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	// Extensions specifies arbitrary metadata of this resource
	Extensions map[string]interface{} `json:"extensions"`
}

func (r *Resource) ResourceKey() string {
	return r.ID
}

func (rs Resources) Index() map[string]*Resource {
	m := make(map[string]*Resource)
	for i := range rs {
		m[rs[i].ResourceKey()] = &rs[i]
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
