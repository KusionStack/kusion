package network

const (
	accessModeInternal = "Internal"
	accessModeExposed  = "Exposed"
	accessProtocolTCP  = "TCP"
	accessProtocolUDP  = "UDP"
)

// ContainerPort defines the available port on the container, and how the other
// containers access the port.
type ContainerPort struct {
	// Port is he available port on the container.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// AccessMode defines the accessibility of the container port.
	AccessMode string `yaml:"accessMode,omitempty" json:"accessMode,omitempty"`

	// AccessPort is the port exposed within the cluster.
	AccessPort int `yaml:"accessPort,omitempty" json:"accessPort, omitempty"`

	// AccessProtocol is the protocol of accessPort.
	AccessProtocol string `yaml:"accessProtocol,omitempty" json:"accessProtocol,omitempty"`
}
