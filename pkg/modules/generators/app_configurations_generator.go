package generators

import (
	"encoding/json"
	"errors"
	"fmt"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload"
	"kusionstack.io/kusion/pkg/modules/proto"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/workspace"
)

type appConfigurationGenerator struct {
	project string
	stack   string
	appName string
	app     *v1.AppConfiguration
	ws      *v1.Workspace
}

func NewAppConfigurationGenerator(
	project string,
	stack string,
	appName string,
	app *v1.AppConfiguration,
	ws *v1.Workspace,
) (modules.Generator, error) {
	if len(project) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(stack) == 0 {
		return nil, fmt.Errorf("stack name must not be empty")
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
		return nil, fmt.Errorf("invalid config of workspace %s, %w", stack, err)
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
	project string,
	stack string,
	appName string,
	app *v1.AppConfiguration,
	ws *v1.Workspace,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewAppConfigurationGenerator(project, stack, appName, app, ws)
	}
}

func (g *appConfigurationGenerator) Generate(spec *v1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(v1.Resources, 0)
	}
	g.app.Name = g.appName

	// retrieve the module configs of the specified project
	projectModuleConfigs, err := workspace.GetProjectModuleConfigs(g.ws.Modules, g.project)
	if err != nil {
		return err
	}

	// todo: is namespace a module? how to retrieve it? Currently, it is configured in the workspace file.
	namespace := g.getNamespaceName(projectModuleConfigs)

	// Generate built-in resources
	gfs := []modules.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(namespace),
		workload.NewWorkloadGeneratorFunc(&workload.Generator{
			Project:         g.project,
			Stack:           g.stack,
			App:             g.appName,
			Namespace:       namespace,
			Workload:        g.app.Workload,
			PlatformConfigs: projectModuleConfigs,
		}),
	}
	if err = modules.CallGenerators(spec, gfs...); err != nil {
		return err
	}

	// Call modules to generate customized resources
	resources, err := g.callModules(projectModuleConfigs)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, resources...)

	// The OrderedResourcesGenerator should be executed after all resources are generated.
	if err = modules.CallGenerators(spec, NewOrderedResourcesGeneratorFunc()); err != nil {
		return err
	}

	// Add kubeConfig from workspace if exist
	modules.AddKubeConfigIf(spec, g.ws)
	return nil
}

func (g *appConfigurationGenerator) callModules(projectModuleConfigs map[string]v1.GenericConfig) ([]v1.Resource, error) {
	var resources []v1.Resource

	pluginMap := make(map[string]*modules.Plugin)
	defer func() {
		for _, plugin := range pluginMap {
			plugin.KillPluginClient()
		}
	}()

	// Generate customized module resources
	for t, config := range projectModuleConfigs {
		// init the plugin
		if pluginMap[t] == nil {
			plugin, err := modules.NewPlugin(t)
			if err != nil {
				return nil, err
			}
			if plugin == nil {
				return nil, fmt.Errorf("init plugin for module %s failed", t)
			}
			pluginMap[t] = plugin
		}
		plugin := pluginMap[t]

		// prepare the request
		devConfig := jsonutil.Marshal2String(g.app.Accessories[t])
		platformConfig := jsonutil.Marshal2String(config)
		wsConfig := jsonutil.Marshal2String(g.ws)
		protoRequest := &proto.GeneratorRequest{
			Project:              g.project,
			Stack:                g.stack,
			App:                  g.appName,
			DevModuleConfig:      []byte(devConfig),
			PlatformModuleConfig: []byte(platformConfig),
			RuntimeConfig:        []byte(wsConfig),
		}

		// invoke the plugin
		response, err := plugin.Module.Generate(protoRequest)
		if err != nil {
			return nil, err
		}
		if response == nil {
			return nil, fmt.Errorf("empty response from module %s", t)
		}
		for _, res := range response.Resources {
			temp := &v1.Resource{}
			err = json.Unmarshal(res, temp)
			if err != nil {
				return nil, err
			}
			// todo: validate resources format in response
			// todo parse patches in the resources
			resources = append(resources, *temp)
		}
	}

	return resources, nil
}

// getNamespaceName obtains the final namespace name using the following precedence
// (from lower to higher):
// - Project name
// - Namespace module config (specified in corresponding workspace file)
func (g *appConfigurationGenerator) getNamespaceName(moduleConfigs map[string]v1.GenericConfig) string {
	if moduleConfigs == nil {
		return g.project
	}

	namespaceName := g.project
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
