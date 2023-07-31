package generators

import (
	"fmt"

	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration"
)

type appConfigurationGenerator struct {
	projectName string
	ac          *appconfiguration.AppConfiguration
}

func NewAppConfigurationGenerator(projectName string, ac *appconfiguration.AppConfiguration) (Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if ac == nil {
		return nil, fmt.Errorf("can not find app configuration when generating the Spec")
	}

	return &appConfigurationGenerator{
		projectName: projectName,
		ac:          ac,
	}, nil
}

func (g *appConfigurationGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	if g.ac.Components != nil {
		gfs := []NewGeneratorFunc{
			NewNamespaceGeneratorFunc(g.projectName),
			NewDeploymentGeneratorFunc(),
			NewJobGeneratorFunc(),
		}

		if err := callGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}
