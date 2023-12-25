package network

const (
	ModulePort = "port"

	FieldType = "type"

	CSPAliCloud = "alicloud"
	CSPAWS      = "aws"
	ProtocolTCP = "TCP"
	ProtocolUDP = "UDP"
)

// Port defines the exposed port of workload.Service, which can be used to describe how
// the workload.Service get accessed.
type Port struct {
	// Type is the specific cloud vendor that provides load balancer, works when Public
	// is true, supports CSPAliCloud and CSPAWS for now.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Port is the exposed port of the workload.Service.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// TargetPort is the backend container.Container port.
	TargetPort int `yaml:"targetPort,omitempty" json:"targetPort,omitempty"`

	// Protocol is protocol used to expose the port, support ProtocolTCP and ProtocolUDP.
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`

	// Public defines whether to expose the port through Internet.
	Public bool `yaml:"public,omitempty" json:"public,omitempty"`
}
