package models

import "encoding/json"

type Type string

type Resources []Resource

type Resource struct {
	// ID is the unique key of this resource in the whole State.
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	ID string `json:"id" yaml:"id"`

	// Type represents all Runtimes we supported
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
