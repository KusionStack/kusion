package workload

import "kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"

// Base defines set of attributes shared by different workload
// profile, e.g. Service and Job. You can inherit this Schema to reuse
// these common attributes.
type Base struct {
	// The templates of containers to be run.
	Containers map[string]container.Container `yaml:"containers,omitempty" json:"containers,omitempty"`

	// The number of containers that should be run.
	// Default is 2 to meet high availability requirements.
	Replicas int `yaml:"replicas,omitempty" json:"replicas,omitempty"`

	// Labels and annotations can be used to attach arbitrary metadata
	// as key-value pairs to resources.
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// Secret
	Secrets map[string]Secret `json:"secrets,omitempty" yaml:"secrets,omitempty"`

	// Dirs configures one or more volumes to be mounted to the
	// specified folder.
	Dirs map[string]string `json:"dirs,omitempty" yaml:"dirs,omitempty"`
}
