package generators

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// SpecGenerator is an interface for things that can generate Spec from input configurations.
// Note: it's for built-in generators to produce Spec, which is not the same as the general Module interface.
type SpecGenerator interface {
	// Generate performs the intent generate operation.
	Generate(intent *v1.Spec) error
}

// NewSpecGeneratorFunc is a function that returns a SpecGenerator.
type NewSpecGeneratorFunc func() (SpecGenerator, error)
