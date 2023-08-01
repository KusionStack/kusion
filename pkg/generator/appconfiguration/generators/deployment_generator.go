package generators

import (
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
)

type deploymentGenerator struct {
	comp *component.Component
}

func NewDeploymentGenerator(comp *component.Component) (Generator, error) {
	return &deploymentGenerator{
		comp: comp,
	}, nil
}

func NewDeploymentGeneratorFunc(comp *component.Component) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewDeploymentGenerator(comp)
	}
}

func (g *deploymentGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	panic("unimplemented")
}
