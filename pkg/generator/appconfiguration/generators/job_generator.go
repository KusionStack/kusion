package generators

import (
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
)

type jobGenerator struct {
	comp *component.Component
}

func NewJobGenerator(comp *component.Component) (Generator, error) {
	return &jobGenerator{
		comp: comp,
	}, nil
}

func NewJobGeneratorFunc(comp *component.Component) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewJobGenerator(comp)
	}
}

func (g *jobGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	panic("unimplemented")
}
