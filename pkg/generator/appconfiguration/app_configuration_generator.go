package appconfiguration

import (
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/appconfiguration/generators"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

type Generator struct {
	*appconfiguration.AppConfiguration
}

func (acg *Generator) GenerateSpec(
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
