package component

import (
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/container"
)

const (
	WorkloadTypeLongRunningService string = "LongRunningService"
	WorkloadTypeJob                string = "Job"
)

type Component struct {
	// The workload type of containers
	WorkloadType string `yaml:"workloadType" json:"workloadType"`
	// The templates of containers to be ran.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`

	// The number of containers that should be ran.
	// Default is 2 to meet high availability requirements.
	// Only supported when workloadType is LongRunningService
	Replicas int `yaml:"replicas,omitempty" json:"replicas,omitempty"`

	// The schedule in Cron format
	// Only supported when workloadType is Job.
	Schedule string

	// List of Workload supporting accessory. Accessory defines various runtime capabilities and operation functionalities.

	// Variables for Day-2 Operation.

	// Variables for Workload scheduling.

	// Other metadata info

	// Labels and annotations can be used to attach arbitrary metadata as key-value pairs to resources.
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}
