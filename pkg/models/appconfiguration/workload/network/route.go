package network

const (
	pathTypeExact  = "Exact"
	pathTypePrefix = "Prefix"
)

// Route enables exposed container ports accessible by HTTP/HTTPS routing paths, then
// the container can be accessed outside the cluster.
// There must be an IngressController in the cluster,so that the configuration of Route
// can work.
type Route struct {
	// Paths is the list of HTTP/HTTPS route paths.
	Paths []RoutePath

	// TLSSecret is the Secret name which contains a TLS private key and certificate.
	TLSSecret string `yaml:"tlsSecret,omitempty" json:"tlsSecret,omitempty"`
}

// RoutePath defines the HTTP/HTTPS path and corresponding backend container accessPort.
type RoutePath struct {
	// The Path exposed to HTTP/HTTPS routing.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// PathType defines how the URL matches path, support empty, PathTypeExact, PathTypePrefix.
	PathType string `yaml:"pathType,omitempty" json:"pathType,omitempty"`

	// ContainerAccessPort is the backend container accessPort.
	ContainerAccessPort string `yaml:"containerAccessPort,omitempty" json:"containerAccessPort,omitempty"`
}
