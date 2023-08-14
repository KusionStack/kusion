package appconfiguration

import (
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/appconfiguration/generators"
	"kusionstack.io/kusion/pkg/models"
	appmodel "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

type Generator struct {
	Apps map[string]appmodel.AppConfiguration
}

func (acg *Generator) GenerateSpec(
	o *generator.Options,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*models.Spec, error) {
	spec := &models.Spec{
		Resources: []models.Resource{},
	}

	gfs := []generators.NewGeneratorFunc{}
	generators.ForeachOrderedApps(acg.Apps, func(appName string, app appmodel.AppConfiguration) error {
		gfs = append(gfs, generators.NewAppConfigurationGeneratorFunc(project.Name, appName, &app))
		return nil
	})
	if err := generators.CallGenerators(spec, gfs...); err != nil {
		return nil, err
	}

	return spec, nil
}
