package generators

import (
	"errors"
	"fmt"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	apisworkspace "kusionstack.io/kusion/pkg/apis/workspace"
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
	project *project.Project
	stack   *stack.Stack
	appName string
	app     *inputs.AppConfiguration
	ws      *apisworkspace.Workspace
}

func NewAppConfigurationGenerator(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	app *inputs.AppConfiguration,
	ws *apisworkspace.Workspace,
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
		return nil, fmt.Errorf("invalid workspace %s, %w", stack.GetName(), err)
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
	project *project.Project,
	stack *stack.Stack,
	appName string,
	app *inputs.AppConfiguration,
	ws *apisworkspace.Workspace,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewAppConfigurationGenerator(project, stack, appName, app, ws)
	}
}

func (g *appConfigurationGenerator) Generate(spec *intent.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(intent.Resources, 0)
	}

	// Generate resources
	gfs := []modules.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(g.project.Name),
		accessories.NewDatabaseGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload, g.app.Database),
		workload.NewWorkloadGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload),
		trait.NewOpsRuleGeneratorFunc(g.project, g.stack, g.appName, g.app),
		monitoring.NewMonitoringGeneratorFunc(g.project, g.app.Monitoring, g.appName),
		// The OrderedResourcesGenerator should be executed after all resources are generated.
		NewOrderedResourcesGeneratorFunc(),
	}
	if err := modules.CallGenerators(spec, gfs...); err != nil {
		return err
	}

	// Patcher logic patches generated resources
	pfs := []modules.NewPatcherFunc{
		pattrait.NewOpsRulePatcherFunc(g.app),
		patmonitoring.NewMonitoringPatcherFunc(g.appName, g.app, g.project),
	}
	if err := modules.CallPatchers(spec.Resources.GVKIndex(), pfs...); err != nil {
		return err
	}

	return nil
}
