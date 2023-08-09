package workload

import "kusionstack.io/kusion/pkg/models/appconfiguration/component/container"

type Job struct {
	// The templates of containers to be ran.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`

	// The schedule in Cron format
	// Only supported when workloadType is Job.
	Schedule string
}
