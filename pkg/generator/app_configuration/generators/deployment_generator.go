package generators

import (
	"kusionstack.io/kusion/pkg/models"
)

type deploymentGenerator struct{}

func NewDeploymentGenerator() (Generator, error) {
	return &deploymentGenerator{}, nil
}

func NewDeploymentGeneratorFunc() NewGeneratorFunc {
	return func() (Generator, error) {
		return NewDeploymentGenerator()
	}
}

func (g *deploymentGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	panic("unimplemented")
}
