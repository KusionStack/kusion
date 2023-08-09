package workload

import "kusionstack.io/kusion/pkg/models/appconfiguration/component/container"

type LongRunningServiceType string

const (
	// LongRunningServiceTypeDeployment is the type of long running service.
	LongRunningServiceTypeDeployment LongRunningServiceType = "Deployment"
	// LongRunningServiceTypeStatefulSet is the type of long running service.
	LongRunningServiceTypeStatefulSet LongRunningServiceType = "StatefulSet"
)

type LongRunningService struct {
	// The type of long running service.
	Type LongRunningServiceType `yaml:"type" json:"type"`

	// The templates of containers to be ran.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`

	// The number of containers that should be ran.
	// Default is 2 to meet high availability requirements.
	// Only supported when workloadType is LongRunningService
	Replicas int `yaml:"replicas,omitempty" json:"replicas,omitempty"`
}
