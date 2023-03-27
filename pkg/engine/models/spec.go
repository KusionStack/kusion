package models

// Spec represents desired state of resources in one stack and will be applied to the actual infrastructure by the Kusion Engine
type Spec struct {
	Resources Resources `json:"resources" yaml:"resources"`
}

// fixme: get `cluster` from compile arguments

// ParseCluster try to parse Cluster from resource extensions.
// All resources in one compile MUST have the same Cluster and this constraint will be guaranteed by KCL compile logic
func (s *Spec) ParseCluster() string {
	resources := s.Resources
	var cluster string
	if len(resources) != 0 && resources[0].Extensions != nil && resources[0].Extensions["Cluster"] != nil {
		cluster = resources[0].Extensions["Cluster"].(string)
	}
	return cluster
}
