package models

// Spec represents desired state of resources in one stack and will be applied to the actual infrastructure by the Kusion Engine
type Spec struct {
	Resources Resources `json:"resources"`
}
