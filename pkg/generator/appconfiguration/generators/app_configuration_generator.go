package generators

import (
	"fmt"

	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration"
)

type appConfigurationGenerator struct {
	projectName string
	appName     string
	app         *appconfiguration.AppConfiguration
}

func NewAppConfigurationGenerator(projectName, appName string, app *appconfiguration.AppConfiguration) (Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if app == nil {
		return nil, fmt.Errorf("can not find app configuration when generating the Spec")
	}

	return &appConfigurationGenerator{
		projectName: projectName,
		appName:     appName,
		app:         app,
	}, nil
}

func NewAppConfigurationGeneratorFunc(projectName, appName string, app *appconfiguration.AppConfiguration) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewAppConfigurationGenerator(projectName, appName, app)
	}
}

func (g *appConfigurationGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	gfs := []NewGeneratorFunc{
		NewNamespaceGeneratorFunc(g.projectName),
		NewWorkloadGeneratorFunc(g.projectName, g.appName, g.app.Workload),
	}

	if err := CallGenerators(spec, gfs...); err != nil {
		return err
	}

	return nil
}
