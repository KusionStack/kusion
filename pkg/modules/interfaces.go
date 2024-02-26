package modules

import (
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// Generator is an interface for things that can generate Intent from input
// configurations.
type Generator interface {
	// Generate performs the intent generate operation.
	Generate(intent *v1.Intent) error
}

// Patcher is the interface that wraps the Patch method.
type Patcher interface {
	Patch(resources map[string][]*v1.Resource) error
}

// NewGeneratorFunc is a function that returns a Generator.
type NewGeneratorFunc func() (Generator, error)

// NewPatcherFunc is a function that returns a Patcher.
type NewPatcherFunc func() (Patcher, error)

// GeneratorRequest defines the request of generators.
type GeneratorRequest struct {
	// Project represents the project name
	Project string
	// Stack represents the stack name
	Stack string
	// App represents the application name
	App string
	// Type represents the module type
	Type string
	// Config is the module inputs of the specific module type
	Config v1.GenericConfig
	// TerraformConfig is the collection of provider configs for the terraform runtime.
	TerraformConfig v1.TerraformConfig
}
