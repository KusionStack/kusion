package v1

// Stack is a definition of Kusion Stack resource.
//
// Stack provides a mechanism to isolate multiple deploys of same application,
// it's the target workspace that an application will be deployed to, also the
// smallest operation unit that can be configured and deployed independently.
type Stack struct {
	// Name is a required fully qualified name.
	Name string `json:"name" yaml:"name"`
	// Description is an optional informational description.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels is the list of labels that are assigned to this stack.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// Path is a directory path within the Git repository.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}
