package generators

import (
	"context"
	"errors"
	"fmt"
	"strings"

	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"kcl-lang.io/kpm/pkg/package"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	internalv1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload"
	"kusionstack.io/kusion/pkg/modules/proto"
	"kusionstack.io/kusion/pkg/workspace"
)

type appConfigurationGenerator struct {
	project      string
	stack        string
	appName      string
	app          *internalv1.AppConfiguration
	ws           *v1.Workspace
	dependencies *pkg.Dependencies
}

var ignoreModules = map[string]bool{
	"service": true,
}

func NewAppConfigurationGenerator(
	project string,
	stack string,
	appName string,
	app *internalv1.AppConfiguration,
	ws *v1.Workspace,
	dependencies *pkg.Dependencies,
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
		return nil, fmt.Errorf("can not find app configuration when generating the Spec")
	}

	if ws == nil {
		return nil, errors.New("workspace must not be empty") // AppConfiguration asks for non-empty workspace
	}

	if err := workspace.ValidateWorkspace(ws); err != nil {
		return nil, fmt.Errorf("invalid config of workspace %s, %w", stack, err)
	}

	return &appConfigurationGenerator{
		project:      project,
		stack:        stack,
		appName:      appName,
		app:          app,
		ws:           ws,
		dependencies: dependencies,
	}, nil
}

func NewAppConfigurationGeneratorFunc(
	project string,
	stack string,
	appName string,
	app *internalv1.AppConfiguration,
	ws *v1.Workspace,
	kpmDependencies *pkg.Dependencies,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewAppConfigurationGenerator(project, stack, appName, app, ws, kpmDependencies)
	}
}

func (g *appConfigurationGenerator) Generate(spec *v1.Spec) error {
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

	// generate built-in resources
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

	// workload is the second generated resource. Check if it is generated.
	if spec.Resources == nil || len(spec.Resources) < 2 {
		return fmt.Errorf("workload is not generated")
	}
	wl := spec.Resources[1]

	// call modules to generate customized resources
	resources, patchers, err := g.callModules(projectModuleConfigs, g.dependencies)
	if err != nil {
		return err
	}

	// patch workload with resource patchers
	for _, p := range patchers {
		if err = PatchWorkload(&wl, &p); err != nil {
			return err
		}
	}

	// append the generated resources to the spec
	spec.Resources = append(spec.Resources, resources...)

	// The OrderedResourcesGenerator should be executed after all resources are generated.
	if err = modules.CallGenerators(spec, NewOrderedResourcesGeneratorFunc()); err != nil {
		return err
	}

	return nil
}

func PatchWorkload(workload *v1.Resource, patcher *internalv1.Patcher) error {
	if patcher == nil {
		return nil
	}

	un := &unstructured.Unstructured{}
	un.SetUnstructuredContent(workload.Attributes)

	// patch labels
	if patcher.Labels != nil {
		objLabels := un.GetLabels()
		if objLabels == nil {
			objLabels = make(map[string]string)
		}
		podLabels, b, err := unstructured.NestedStringMap(un.Object, "spec", "template", "metadata", "labels")
		if err != nil {
			return fmt.Errorf("failed to get pod labels from workload:%s. %w", workload.ID, err)
		}
		if !b || podLabels == nil {
			podLabels = make(map[string]string)
		}
		// merge labels
		for k, v := range patcher.Labels {
			objLabels[k] = v
			podLabels[k] = v
		}
		un.SetLabels(objLabels)
		err = unstructured.SetNestedStringMap(un.Object, podLabels, "spec", "template", "metadata", "labels")
		if err != nil {
			return err
		}
	}

	// patch annotations
	if patcher.Annotations != nil {
		objAnnotations := un.GetAnnotations()
		if objAnnotations == nil {
			objAnnotations = make(map[string]string)
		}
		podAnnotations, b, err := unstructured.NestedStringMap(un.Object, "spec", "template", "metadata", "annotations")
		if err != nil {
			return fmt.Errorf("failed to get pod annotations from workload:%s. %w", workload.ID, err)
		}
		if !b || podAnnotations == nil {
			podAnnotations = make(map[string]string)
		}
		// merge annotations
		for k, v := range patcher.Annotations {
			objAnnotations[k] = v
			podAnnotations[k] = v
		}

		un.SetAnnotations(objAnnotations)
		err = unstructured.SetNestedStringMap(un.Object, podAnnotations, "spec", "template", "metadata", "annotations")
		if err != nil {
			return err
		}
	}

	// patch env
	if patcher.Environments != nil {
		containers, b, err := unstructured.NestedSlice(un.Object, "spec", "template", "spec", "containers")
		if err != nil || !b {
			return fmt.Errorf("failed to get containers from workload:%s. %w", workload.ID, err)
		}
		// merge env
		for i, c := range containers {
			container := c.(map[string]interface{})
			envs, b, err := unstructured.NestedSlice(container, "env")
			if err != nil {
				return fmt.Errorf("failed to get env from workload:%s, container:%s. %w", workload.ID, container["name"], err)
			}
			if !b {
				envs = make([]interface{}, 0)
			}

			for _, env := range patcher.Environments {
				us, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&env)
				if err != nil {
					return err
				}
				// prepend patch env to existing env slices so developers can reference them later on
				// ref: https://kubernetes.io/docs/tasks/inject-data-application/define-interdependent-environment-variables/
				envs = append([]interface{}{us}, envs...)
				log.Info("we're gonna patch env:%s,value:%s to workload:%s, container:%s", env.Name, env.Value, workload.ID, container["name"])
			}

			container["env"] = envs
			containers[i] = container
		}

		if err = unstructured.SetNestedSlice(un.Object, containers, "spec", "template", "spec", "containers"); err != nil {
			return err
		}
	}

	return nil
}

// moduleConfig represents the configuration of a module, either devConfig or platformConfig can be nil
type moduleConfig struct {
	devConfig      internalv1.Accessory
	platformConfig v1.GenericConfig
}

func (g *appConfigurationGenerator) callModules(
	projectModuleConfigs map[string]v1.GenericConfig, dependencies *pkg.Dependencies,
) (resources []v1.Resource, patchers []internalv1.Patcher, err error) {
	pluginMap := make(map[string]*modules.Plugin)
	defer func() {
		for _, plugin := range pluginMap {
			pluginErr := plugin.KillPluginClient()
			if pluginErr != nil {
				err = fmt.Errorf("call modules failed %w. %s", err, pluginErr)
			}
		}
	}()

	// build module config index
	if dependencies == nil {
		return nil, nil, errors.New("dependencies should not be nil")
	}
	indexModuleConfig, err := buildModuleConfigIndex(g.app.Accessories, projectModuleConfigs, dependencies)
	if err != nil {
		return nil, nil, err
	}

	// generate customized module resources
	for t, config := range indexModuleConfig {
		// ignore workload and namespace modules
		if ignoreModules[t] {
			continue
		}

		// init the plugin
		if pluginMap[t] == nil {
			plugin, err := modules.NewPlugin(t)
			if err != nil {
				return nil, nil, err
			}
			if plugin == nil {
				return nil, nil, fmt.Errorf("init plugin for module %s failed", t)
			}
			pluginMap[t] = plugin
		}
		plugin := pluginMap[t]

		// prepare the request
		protoRequest, err := g.initModuleRequest(config)
		if err != nil {
			return nil, nil, err
		}

		// invoke the plugin
		log.Infof("invoke module:%s with request:%s", t, protoRequest.String())
		response, err := plugin.Module.Generate(context.Background(), protoRequest)
		if err != nil {
			return nil, nil, fmt.Errorf("invoke kusion module: %s failed. %w", t, err)
		}
		if response == nil {
			return nil, nil, fmt.Errorf("empty response from module %s", t)
		}

		// parse module result
		for _, res := range response.Resources {
			temp := &v1.Resource{}
			err = yaml.Unmarshal(res, temp)
			if err != nil {
				return nil, nil, err
			}
			resources = append(resources, *temp)
		}

		// parse patcher
		for _, patcher := range response.Patchers {
			temp := &internalv1.Patcher{}
			err = yaml.Unmarshal(patcher, temp)
			if err != nil {
				return nil, nil, err
			}
			patchers = append(patchers, *temp)
		}
	}

	return resources, patchers, nil
}

func buildModuleConfigIndex(
	accessories map[string]internalv1.Accessory,
	projectModuleConfigs map[string]v1.GenericConfig,
	dependencies *pkg.Dependencies,
) (map[string]moduleConfig, error) {
	indexModuleConfig := map[string]moduleConfig{}
	for accName, accessory := range accessories {
		// parse accessory module key
		key, err := parseModuleKey(accessory, dependencies)
		if err != nil {
			return nil, err
		}
		log.Info("build module index of accessory:%s module key: %s", accName, key)
		indexModuleConfig[key] = moduleConfig{
			devConfig:      accessory,
			platformConfig: projectModuleConfigs[key],
		}
	}
	// append module configs only exist in platform configs
	for key, platformConfig := range projectModuleConfigs {
		if _, ok := indexModuleConfig[key]; !ok {
			indexModuleConfig[key] = moduleConfig{
				devConfig:      nil,
				platformConfig: platformConfig,
			}
		}
	}
	return indexModuleConfig, nil
}

func parseModuleKey(accessory internalv1.Accessory, dependencies *pkg.Dependencies) (string, error) {
	split := strings.Split(accessory["_type"].(string), ".")
	moduleName := split[0]
	// find module namespace and version
	d, ok := dependencies.Deps[moduleName]
	if !ok {
		return "", fmt.Errorf("can not find module %s in dependencies", moduleName)
	}
	// key example "kusionstack/mysql@v0.1.0"
	var key string
	if d.Oci != nil {
		key = fmt.Sprintf("%s@%s", d.Oci.Repo, d.Version)
	} else if d.Git != nil {
		// todo: kpm will change the repo version with the filed `version` in d.Git.Version
		url := strings.TrimSuffix(d.Git.Url, ".git")
		splits := strings.Split(url, "/")
		ns := splits[len(splits)-2] + "/" + splits[len(splits)-1]
		key = fmt.Sprintf("%s@%s", ns, d.Git.Tag)
	}
	return key, nil
}

func (g *appConfigurationGenerator) initModuleRequest(config moduleConfig) (*proto.GeneratorRequest, error) {
	var workloadConfig, devConfig, platformConfig []byte
	var err error
	// Attention: we MUST yaml.v2 to serialize the object,
	// because we have introduced MapSlice in the Workload which is supported only in the yaml.v2
	if g.app.Workload != nil {
		if workloadConfig, err = yamlv2.Marshal(g.app.Workload); err != nil {
			return nil, fmt.Errorf("marshal workload config failed. %w", err)
		}
	}
	if config.devConfig != nil {
		if devConfig, err = yaml.Marshal(config.devConfig); err != nil {
			return nil, fmt.Errorf("marshal dev module config failed. %w", err)
		}
	}
	if config.platformConfig != nil {
		if platformConfig, err = yaml.Marshal(config.platformConfig); err != nil {
			return nil, fmt.Errorf("marshal platform module config failed. %w", err)
		}
	}

	protoRequest := &proto.GeneratorRequest{
		Project:              g.project,
		Stack:                g.stack,
		App:                  g.appName,
		Workload:             workloadConfig,
		DevModuleConfig:      devConfig,
		PlatformModuleConfig: platformConfig,
	}
	return protoRequest, nil
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
