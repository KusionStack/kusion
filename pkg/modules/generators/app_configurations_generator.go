package generators

import (
	"errors"
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	accessories "kusionstack.io/kusion/pkg/modules/generators/accessories/database"
	"kusionstack.io/kusion/pkg/modules/generators/monitoring"
	"kusionstack.io/kusion/pkg/modules/generators/trait"
	"kusionstack.io/kusion/pkg/modules/generators/workload"
	"kusionstack.io/kusion/pkg/modules/inputs"
	patmonitoring "kusionstack.io/kusion/pkg/modules/patchers/monitoring"
	pattrait "kusionstack.io/kusion/pkg/modules/patchers/trait"
	"kusionstack.io/kusion/pkg/workspace"
)

type appConfigurationGenerator struct {
	project *apiv1.Project
	stack   *apiv1.Stack
	appName string
	app     *inputs.AppConfiguration
	ws      *apiv1.Workspace
}

func NewAppConfigurationGenerator(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	app *inputs.AppConfiguration,
	ws *apiv1.Workspace,
) (modules.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if app == nil {
		return nil, fmt.Errorf("can not find app configuration when generating the Intent")
	}

	if ws == nil {
		return nil, errors.New("workspace must not be empty") // AppConfiguration asks for non-empty workspace
	}
	if err := workspace.ValidateWorkspace(ws); err != nil {
		return nil, fmt.Errorf("invalid config of workspace %s, %w", stack.Name, err)
	}

	return &appConfigurationGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
		ws:      ws,
	}, nil
}

func NewAppConfigurationGeneratorFunc(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	app *inputs.AppConfiguration,
	ws *apiv1.Workspace,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewAppConfigurationGenerator(project, stack, appName, app, ws)
	}
}

func (g *appConfigurationGenerator) Generate(i *apiv1.Intent) error {
	if i.Resources == nil {
		i.Resources = make(apiv1.Resources, 0)
	}

	// retrieve the module configs of the specified project
	modulesConfig, err := workspace.GetProjectModuleConfigs(g.ws.Modules, g.project.Name)
	if err != nil {
		return err
	}

	// Generate resources
	gfs := []modules.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(g.project.Name, g.ws),
		accessories.NewDatabaseGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload, g.app.Database),
		workload.NewWorkloadGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload, moduleConfigs),
		trait.NewOpsRuleGeneratorFunc(g.project, g.stack, g.appName, g.app, modulesConfig),
		monitoring.NewMonitoringGeneratorFunc(g.project, g.app.Monitoring, g.appName),
		// The OrderedResourcesGenerator should be executed after all resources are generated.
		NewOrderedResourcesGeneratorFunc(),
	}
	if err := modules.CallGenerators(i, gfs...); err != nil {
		return err
	}

	// Patcher logic patches generated resources
	pfs := []modules.NewPatcherFunc{
		pattrait.NewOpsRulePatcherFunc(g.app, modulesConfig),
		patmonitoring.NewMonitoringPatcherFunc(g.appName, g.app, g.project),
	}
	if err := modules.CallPatchers(i.Resources.GVKIndex(), pfs...); err != nil {
		return err
	}

	// Add kubeConfig from workspace if exist
	modules.AddKubeConfigIf(i, g.ws)

	return nil
}
