package v1

type Type string

const (
	Kubernetes Type = "Kubernetes"
	Terraform  Type = "Terraform"
)

// Intent describes the desired state how the infrastructure should look like: which workload to run,
// the load-balancer setup, the location of the database schema, and so on. Based on that information,
// the Kusion engine takes care of updating the production state to match the Intent.
type Intent struct {
	// Resources is the list of Resource this Intent contains.
	Resources Resources `json:"resources" yaml:"resources"`
}

type Resources []Resource

// Resource is the representation of a resource in the state.
type Resource struct {
	// ID is the unique key of this resource in the whole State.
	// ApiVersion:Kind:Namespace:Name is an idiomatic way for Kubernetes resources.
	// providerNamespace:providerName:resourceType:resourceName for Terraform resources
	ID string `json:"id" yaml:"id"`

	// Type represents all Runtimes we supported like Kubernetes and Terraform
	Type Type `json:"type" yaml:"type"`

	// Attributes represents all specified attributes of this resource
	Attributes map[string]interface{} `json:"attributes" yaml:"attributes"`

	// DependsOn contains all resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	// Extensions specifies arbitrary metadata of this resource
	Extensions map[string]interface{} `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}
