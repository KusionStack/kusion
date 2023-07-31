package component

import (
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/container"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/job"
)

type Component struct {
	// The templates of containers to be ran.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`
	// The templates of jobs to be ran.
	Jobs map[string]job.Job `yaml:"jobs,omitempty" json:"jobs,omitempty"`

	// The number of containers that should be ran.
	// Default is 2 to meet high availability requirements.
	Replicas int `yaml:"replicas,omitempty" json:"replicas,omitempty"`

	// List of Workload supporting accessory. Accessory defines various runtime capabilities and operation functionalities.

	// Variables for Day-2 Operation.

	// Variables for Workload scheduling.

	// Other metadata info

	// Labels and annotations can be used to attach arbitrary metadata as key-value pairs to resources.
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}
