package workload

import (
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
)

// Service is a kind of workload profile that describes how to run
// your application code. This is typically used for long-running web
// applications that should "never" go down, and handle short-lived
// latency-sensitive web requests, or events.
type Service struct {
	WorkloadBase `yaml:",inline" json:",inline"`

	// Routes contains the hostnames and corresponding HTTP/HTTPS routes to
	// expose the Service outside the cluster.
	Routes map[string]network.Route `yaml:"routes,omitempty" json:"routes,omitempty"`
}
