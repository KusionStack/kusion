package generator

import (
	"fmt"

	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/generator/appconfiguration/generator/workload"
	"kusionstack.io/kusion/pkg/models"
	appmodel "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

type AppsGenerator struct {
	Apps map[string]appmodel.AppConfiguration
}

func (acg *AppsGenerator) GenerateSpec(
	o *generator.Options,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*models.Spec, error) {
	spec := &models.Spec{
		Resources: []models.Resource{},
	}

	gfs := []appconfiguration.NewGeneratorFunc{}
	appconfiguration.ForeachOrdered(acg.Apps, func(appName string, app appmodel.AppConfiguration) error {
		gfs = append(gfs, NewAppConfigurationGeneratorFunc(project.Name, appName, &app))
		return nil
	})
	if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
		return nil, err
	}

	return spec, nil
}

type appConfigurationGenerator struct {
	projectName string
	appName     string
	app         *appmodel.AppConfiguration
}

func NewAppConfigurationGenerator(
	projectName, appName string,
	app *appmodel.AppConfiguration,
) (appconfiguration.Generator, error) {
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

func NewAppConfigurationGeneratorFunc(
	projectName, appName string,
	app *appmodel.AppConfiguration,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewAppConfigurationGenerator(projectName, appName, app)
	}
}

func (g *appConfigurationGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	gfs := []appconfiguration.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(g.projectName),
		workload.NewWorkloadGeneratorFunc(g.projectName, g.appName, g.app.Workload),
	}

	if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
		return err
	}

	return nil
}
