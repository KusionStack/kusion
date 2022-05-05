package manifest

import "kusionstack.io/kusion/pkg/engine/states"

// Manifest represent the KCL compile result
type Manifest struct {
	Resources states.Resources `json:"resources"`
}
