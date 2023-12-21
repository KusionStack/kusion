package v1

type (
	BuilderType string
	MonitorType string
)

// Project is a definition of Kusion Project resource.
//
// A project is composed of one or more applications and is linked to a Git repository,
// which contains the project's desired manifests.
type Project struct {
	// Name is a required fully qualified name.
	Name string `json:"name" yaml:"name"`

	// Description is an optional informational description.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels is the list of labels that are assigned to this project.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Generator controls how to generate the Intent.
	Generator *GeneratorConfig `json:"generator,omitempty" yaml:"generator,omitempty"`
}

// GeneratorConfig holds the intent generation configurations defined in Project resource.
type GeneratorConfig struct {
	// Type specifies the type of Generator. can be either "KCL" or "AppConfiguration".
	Type BuilderType `json:"type" yaml:"type"`
	// Configs contains extra configurations used by the Generator.
	Configs map[string]interface{} `json:"configs,omitempty" yaml:"configs,omitempty"`
}

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
}
