package workload

import (
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
)

// WorkloadBase defines set of attributes shared by different workload
// profile, e.g. Service and Job. You can inherit this Schema to reuse
// these common attributes.
type WorkloadBase struct {
	// The templates of containers to be run.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`

	// The number of containers that should be run.
	// Default is 2 to meet high availability requirements.
	Replicas int `yaml:"replicas,omitempty" json:"replicas,omitempty"`

	// Labels and annotations can be used to attach arbitrary metadata
	// as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// Dirs configures one or more volumes to be mounted to the
	// specified folder.
	Dirs map[string]string `json:"dirs,omitempty" yaml:"dirs,omitempty"`
	// Files configures one or more files to be created in the container.
	// files                     {str:FileSpec}

	// Liveness probe for this container.
	// Liveness probe indicates if a running process is healthy.
	// livenessProbe             p.Probe`json:"livenessProbe,omitempty" yaml:"livenessProbe,omitempty"`

	// Readiness probe for this container.
	// Readiness probe indicates whether an application is available
	// to handle requests.
	// readinessProbe            p.Probe`json:"readinessProbe,omitempty" yaml:"readinessProbe,omitempty"`

	// Startup probe for this container.
	// Startup probe indicates that the container has started for the
	// first time.
	// startupProbe              p.Probe`json:"startupProbe,omitempty" yaml:"startupProbe,omitempty"`

	// Lifecycle configures actions which should be taken response to
	// container lifecycle events.
	// lifecycle                 lc.Lifecycle`json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
}
