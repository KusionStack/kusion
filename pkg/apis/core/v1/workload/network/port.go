package network

type Protocol string

const (
	TCP Protocol = "TCP"
	UDP Protocol = "UDP"
)

// Port defines the exposed port of workload.Service
type Port struct {
	// Port is the exposed port of the workload.Service.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
	// TargetPort is the backend container.Container port.
	TargetPort int `yaml:"targetPort,omitempty" json:"targetPort,omitempty"`
	// Protocol is protocol used to expose the port, support ProtocolTCP and ProtocolUDP.
	Protocol Protocol `yaml:"protocol,omitempty" json:"protocol,omitempty"`
}
