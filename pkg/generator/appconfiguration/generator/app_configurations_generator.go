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
	_ *generator.Options,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*models.Spec, error) {
	spec := &models.Spec{
		Resources: []models.Resource{},
	}

	var gfs []appconfiguration.NewGeneratorFunc
	err := appconfiguration.ForeachOrdered(acg.Apps, func(appName string, app appmodel.AppConfiguration) error {
		gfs = append(gfs, NewAppConfigurationGeneratorFunc(project, stack, appName, &app))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
		return nil, err
	}

	return spec, nil
}

type appConfigurationGenerator struct {
	project *projectstack.Project
	stack   *projectstack.Stack
	appName string
	app     *appmodel.AppConfiguration
}

func NewAppConfigurationGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	app *appmodel.AppConfiguration,
	appName string,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if app == nil {
		return nil, fmt.Errorf("can not find app configuration when generating the Spec")
	}

	return &appConfigurationGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
	}, nil
}

func NewAppConfigurationGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	app *appmodel.AppConfiguration,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewAppConfigurationGenerator(project, stack, app, appName)
	}
}

func (g *appConfigurationGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	gfs := []appconfiguration.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(g.project.Name),
		workload.NewWorkloadGeneratorFunc(g.project, nil, g.app.Workload, g.appName),
	}

	if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
		return err
	}

	return nil
}
