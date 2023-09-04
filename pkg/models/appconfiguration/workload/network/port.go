package network

const (
	ProtocolTCP = "TCP"
	ProtocolUDP = "UDP"
)

// Port defines the exposed port of workload.Service, which can be used to describe how
// the workload.Service get accessed.
type Port struct {
	// Port is the exposed port of the workload.Service.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// TargetPort is the backend container.Container port.
	TargetPort int `yaml:"targetPort,omitempty" json:"targetPort,omitempty"`

	// Protocol is protocol used to expose the port, support ProtocolTCP and ProtocolUDP.
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`

	// Public defines whether to expose the port through Internet.
	Public bool `yaml:"public,omitempty" json:"public,omitempty"`
}
