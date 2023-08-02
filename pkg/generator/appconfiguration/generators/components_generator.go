package generators

import (
	"fmt"

	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
)

type componentsGenerator struct {
	projectName string
	components  map[string]component.Component
}

func NewComponentsGenerator(projectName string, components map[string]component.Component) (Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &componentsGenerator{
		projectName: projectName,
		components:  components,
	}, nil
}

func NewComponentsGeneratorFunc(projectName string, components map[string]component.Component) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewComponentsGenerator(projectName, components)
	}
}

func (g *componentsGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	if g.components != nil {
		for compName, comp := range g.components {
			gfs := []NewGeneratorFunc{
				NewDeploymentGeneratorFunc(g.projectName, compName, &comp),
			}

			if err := callGenerators(spec, gfs...); err != nil {
				return err
			}
		}
	}

	return nil
}
