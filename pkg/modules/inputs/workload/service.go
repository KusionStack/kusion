package workload

import (
	"kusionstack.io/kusion/pkg/modules/inputs/workload/network"
)

const (
	TypeDeployment = "Deployment"
	TypeCollaset   = "CollaSet"
)

// Service is a kind of workload profile that describes how to run
// your application code. This is typically used for long-running web
// applications that should "never" go down, and handle short-lived
// latency-sensitive web requests, or events.
type Service struct {
	Base `yaml:",inline" json:",inline"`
	Type string `yaml:"type" json:"type"`

	// Ports describe the list of ports need getting exposed.
	Ports []network.Port `yaml:"ports,omitempty" json:"ports,omitempty"`
}
