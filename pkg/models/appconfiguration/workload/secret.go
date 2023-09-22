package workload

import v1 "k8s.io/api/core/v1"

type Secret struct {
	// Image to run for this container
	Type      v1.SecretType
	Data      map[string]string `yaml:"data,omitempty" json:"data,omitempty"`
	Immutable bool              `yaml:"immutable,omitempty" json:"immutable,omitempty"`
}
