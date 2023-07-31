package app_configuration

import (
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/app_configuration/generators"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

type AppConfigurationGenerator struct {
	*appconfiguration.AppConfiguration
}

func (acg *AppConfigurationGenerator) GenerateSpec(
	o *generator.Options,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*models.Spec, error) {
	spec := &models.Spec{
		Resources: []models.Resource{},
	}

	g, err := generators.NewAppConfigurationGenerator(project.Name, acg.AppConfiguration)
	if err != nil {
		return nil, err
	}
	if err = g.Generate(spec); err != nil {
		return nil, err
	}

	return spec, nil
}
