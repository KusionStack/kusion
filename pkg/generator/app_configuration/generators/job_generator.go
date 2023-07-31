package generators

import (
	"kusionstack.io/kusion/pkg/models"
)

type jobGenerator struct{}

func NewJobGenerator() (Generator, error) {
	return &jobGenerator{}, nil
}

func NewJobGeneratorFunc() NewGeneratorFunc {
	return func() (Generator, error) {
		return NewJobGenerator()
	}
}

func (g *jobGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	panic("unimplemented")
}
