package generators

import (
	"errors"
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload"
	"kusionstack.io/kusion/pkg/workspace"
)

type appConfigurationGenerator struct {
	project *apiv1.Project
	stack   *apiv1.Stack
	appName string
	app     *apiv1.AppConfiguration
	ws      *apiv1.Workspace
}

func NewAppConfigurationGenerator(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	app *apiv1.AppConfiguration,
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
	app *apiv1.AppConfiguration,
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
	g.app.Name = g.appName

	// retrieve the module configs of the specified project
	platformConfigs, err := workspace.GetProjectModuleConfigs(g.ws.Modules, g.project.Name)
	if err != nil {
		return err
	}

	// todo: is namespace a module? how to retrieve it? Currently, it is configured in the workspace file.
	namespace := g.getNamespaceName(platformConfigs)

	// Generate built-in resources
	gfs := []modules.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(namespace),
		workload.NewWorkloadGeneratorFunc(&workload.Generator{
			Project:         g.project.Name,
			Stack:           g.stack.Name,
			App:             g.appName,
			Namespace:       namespace,
			Workload:        g.app.Workload,
			PlatformConfigs: platformConfigs,
		}),
	}
	if err = modules.CallGenerators(i, gfs...); err != nil {
		return err
	}

	// Generate customized module resources
	for range platformConfigs {
		// todo: invoke kusion module generators for each module
	}

	// The OrderedResourcesGenerator should be executed after all resources are generated.
	if err := modules.CallGenerators(i, NewOrderedResourcesGeneratorFunc()); err != nil {
		return err
	}

	// Add kubeConfig from workspace if exist
	modules.AddKubeConfigIf(i, g.ws)

	return nil
}

// getNamespaceName obtains the final namespace name using the following precedence
// (from lower to higher):
// - Project name
// - Namespace module config (specified in corresponding workspace file)
func (g *appConfigurationGenerator) getNamespaceName(moduleConfigs map[string]apiv1.GenericConfig) string {
	if moduleConfigs == nil {
		return g.project.Name
	}

	namespaceName := g.project.Name
	namespaceModuleConfigs, exist := moduleConfigs["namespace"]
	if exist {
		if name, ok := namespaceModuleConfigs["name"]; ok {
			customNamespaceName, isString := name.(string)
			if isString && len(customNamespaceName) > 0 {
				namespaceName = customNamespaceName
			}
		}
	}
	return namespaceName
}
