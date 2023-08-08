package workload

import "kusionstack.io/kusion/pkg/models/appconfiguration/component/container"

type LongRunningService struct {
	// The templates of containers to be ran.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`

	// The number of containers that should be ran.
	// Default is 2 to meet high availability requirements.
	// Only supported when workloadType is LongRunningService
	Replicas int `yaml:"replicas,omitempty" json:"replicas,omitempty"`
}
