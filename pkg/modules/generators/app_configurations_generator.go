// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generators

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/util/json"
	"kcl-lang.io/kpm/pkg/package"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload"
	"kusionstack.io/kusion/pkg/modules/proto"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/workspace"
)

type appConfigurationGenerator struct {
	project      *v1.Project
	stack        *v1.Stack
	appName      string
	app          *v1.AppConfiguration
	ws           *v1.Workspace
	dependencies *pkg.Dependencies
}

func NewAppConfigurationGenerator(
	project *v1.Project,
	stack *v1.Stack,
	appName string,
	app *v1.AppConfiguration,
	ws *v1.Workspace,
	dependencies *pkg.Dependencies,
) (modules.Generator, error) {
	if project == nil {
		return nil, fmt.Errorf("project must not be nil")
	}

	if stack == nil {
		return nil, fmt.Errorf("stack must not be nil")
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
		return nil, fmt.Errorf("invalid config of workspace: %s, %w", ws.Name, err)
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
	project *v1.Project,
	stack *v1.Stack,
	appName string,
	app *v1.AppConfiguration,
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
	projectModuleConfigs, err := workspace.GetProjectModuleConfigs(g.ws.Modules, g.project.Name)
	if err != nil {
		return err
	}

	// generate built-in resources
	namespace := g.getNamespaceName()
	gfs := []modules.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(namespace),
		workload.NewWorkloadGeneratorFunc(&workload.Generator{
			Project:         g.project.Name,
			Stack:           g.stack.Name,
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
	resources, patcher, err := g.callModules(projectModuleConfigs)
	if err != nil {
		return err
	}

	// append the generated resources to the spec
	spec.Resources = append(spec.Resources, resources...)

	// patch workload with resource patcher
	if patcher != nil {
		if err = PatchWorkload(&wl, patcher); err != nil {
			return err
		}
		if err = JSONPatch(spec.Resources, patcher); err != nil {
			return err
		}
	}

	// The OrderedResourcesGenerator should be executed after all resources are generated.
	if err = modules.CallGenerators(spec, NewOrderedResourcesGeneratorFunc()); err != nil {
		return err
	}

	return nil
}

func JSONPatch(resources v1.Resources, patcher *v1.Patcher) error {
	if resources == nil || patcher == nil {
		return nil
	}

	resIndex := resources.Index()

	if patcher.JSONPatchers != nil {
		for id, jsonPatcher := range patcher.JSONPatchers {
			res, ok := resIndex[id]
			if !ok {
				return fmt.Errorf("target patch resource %s not found", id)
			}

			target := jsonutil.Marshal2String(res.Attributes)
			switch jsonPatcher.Type {
			case v1.MergePatch:
				modified, err := jsonpatch.MergePatch([]byte(target), jsonPatcher.Payload)
				if err != nil {
					return fmt.Errorf("merge patch to:%s failed", id)
				}
				if err = json.Unmarshal(modified, &res.Attributes); err != nil {
					return err
				}
			case v1.JSONPatch:
				patch, err := jsonpatch.DecodePatch(jsonPatcher.Payload)
				if err != nil {
					return fmt.Errorf("decode json patch:%s failed", jsonPatcher.Payload)
				}

				modified, err := patch.Apply([]byte(target))
				if err != nil {
					return fmt.Errorf("apply json patch to:%s failed", id)
				}
				if err = json.Unmarshal(modified, &res.Attributes); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported patch type:%s", jsonPatcher.Type)
			}
		}
	}
	return nil
}

func PatchWorkload(workload *v1.Resource, patcher *v1.Patcher) error {
	if patcher == nil {
		return nil
	}

	un := &unstructured.Unstructured{}
	attributes := workload.Attributes

	// normalize attributes with K8s json util. Especially numbers are converted to int64 or float64
	out, err := k8sjson.Marshal(attributes)
	if err != nil {
		return err
	}
	if err = k8sjson.Unmarshal(out, &attributes); err != nil {
		return err
	}
	un.SetUnstructuredContent(attributes)

	// patch labels
	if patcher.Labels != nil {
		objLabels := un.GetLabels()
		if objLabels == nil {
			objLabels = make(map[string]string)
		}
		// merge labels
		for k, v := range patcher.Labels {
			objLabels[k] = v
		}
		un.SetLabels(objLabels)
	}

	// patch pod labels
	if patcher.PodLabels != nil {
		podLabels, b, err := unstructured.NestedStringMap(un.Object, "spec", "template", "metadata", "labels")
		if err != nil {
			return fmt.Errorf("failed to get pod labels from workload:%s. %w", workload.ID, err)
		}
		if !b || podLabels == nil {
			podLabels = make(map[string]string)
		}
		// merge labels
		for k, v := range patcher.PodLabels {
			podLabels[k] = v
		}
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
		// merge annotations
		for k, v := range patcher.Annotations {
			objAnnotations[k] = v
		}
		un.SetAnnotations(objAnnotations)
	}

	// patch pod annotations
	if patcher.PodAnnotations != nil {
		podAnnotations, b, err := unstructured.NestedStringMap(un.Object, "spec", "template", "metadata", "annotations")
		if err != nil {
			return fmt.Errorf("failed to get pod annotations from workload:%s. %w", workload.ID, err)
		}
		if !b || podAnnotations == nil {
			podAnnotations = make(map[string]string)
		}
		// merge annotations
		for k, v := range patcher.PodAnnotations {
			podAnnotations[k] = v
		}
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
	devConfig      v1.Accessory
	platformConfig v1.GenericConfig
	ctx            v1.GenericConfig
}

func (g *appConfigurationGenerator) callModules(projectModuleConfigs map[string]v1.GenericConfig) (resources []v1.Resource, patcher *v1.Patcher, err error) {
	pluginMap := make(map[string]*modules.Plugin)
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf("call modules panic:%s", e)
			case error:
				err = x
			default:
				err = errors.New("call modules unknown panic")
			}
		}
		for _, plugin := range pluginMap {
			pluginErr := plugin.KillPluginClient()
			if pluginErr != nil {
				err = fmt.Errorf("kill modules failed %w. %s", err, pluginErr)
			}
		}
		if err != nil {
			log.Errorf(err.Error())
		}
	}()

	// build module config index
	if g.dependencies == nil {
		return nil, nil, errors.New("dependencies should not be nil")
	}
	indexModuleConfig, err := g.buildModuleConfigIndex(projectModuleConfigs)
	if err != nil {
		return nil, nil, err
	}

	// generate customized module resources
	for t, config := range indexModuleConfig {
		// ignore workload modules
		if modules.IgnoreModules[t] {
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
		err = yaml.Unmarshal(response.Patcher, &patcher)
		if err != nil {
			return nil, nil, err
		}
	}

	return resources, patcher, nil
}

func (g *appConfigurationGenerator) buildModuleConfigIndex(platformModuleConfigs map[string]v1.GenericConfig) (map[string]moduleConfig, error) {
	indexModuleConfig := map[string]moduleConfig{}
	for accName, accessory := range g.app.Accessories {
		// parse accessory module key
		key, err := parseModuleKey(accessory, g.dependencies)
		if err != nil {
			return nil, err
		}
		log.Info("build module index of accessory:%s module key: %s", accName, key)
		moduleName, err := getModuleName(accessory)
		if err != nil {
			return nil, err
		}
		indexModuleConfig[key] = moduleConfig{
			devConfig:      accessory,
			platformConfig: platformModuleConfigs[moduleName],
			ctx:            g.ws.Context,
		}
	}
	return indexModuleConfig, nil
}

// parseModuleKey returns the module key of the accessory in format of "org/module@version"
// example: "kusionstack/mysql@v0.1.0"
func parseModuleKey(accessory v1.Accessory, dependencies *pkg.Dependencies) (string, error) {
	moduleName, err := getModuleName(accessory)
	if err != nil {
		return "", err
	}

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

func getModuleName(accessory v1.Accessory) (string, error) {
	t, ok := accessory["_type"]
	if !ok {
		return "", errors.New("can not find '_type' in module config")
	}
	split := strings.Split(t.(string), ".")
	return split[0], nil
}

func (g *appConfigurationGenerator) initModuleRequest(config moduleConfig) (*proto.GeneratorRequest, error) {
	var workloadConfig, devConfig, platformConfig, ctx []byte
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
	if config.ctx != nil {
		if ctx, err = yaml.Marshal(config.ctx); err != nil {
			return nil, fmt.Errorf("marshal context config failed. %w", err)
		}
	}

	protoRequest := &proto.GeneratorRequest{
		Project:        g.project.Name,
		Stack:          g.stack.Name,
		App:            g.appName,
		Workload:       workloadConfig,
		DevConfig:      devConfig,
		PlatformConfig: platformConfig,
		Context:        ctx,
	}
	return protoRequest, nil
}

// getNamespaceName obtains the final namespace name using the following precedence
// (from lower to higher):
// - Project name
// - KubernetesNamespace extensions (specified in corresponding workspace file)
func (g *appConfigurationGenerator) getNamespaceName() string {
	extensions := mergeExtensions(g.project, g.stack)
	if len(extensions) != 0 {
		for _, extension := range extensions {
			switch extension.Kind {
			case v1.KubernetesNamespace:
				return extension.KubeNamespace.Namespace
			default:
				// do nothing
			}
		}
	}

	return g.project.Name
}

func mergeExtensions(project *v1.Project, stack *v1.Stack) []*v1.Extension {
	var extensions []*v1.Extension
	extensionKindMap := make(map[string]struct{})
	if stack.Extensions != nil && len(stack.Extensions) != 0 {
		for _, extension := range stack.Extensions {
			extensions = append(extensions, extension)
			extensionKindMap[string(extension.Kind)] = struct{}{}
		}
	}
	if project.Extensions != nil && len(project.Extensions) != 0 {
		for _, extension := range project.Extensions {
			if _, exist := extensionKindMap[string(extension.Kind)]; !exist {
				extensions = append(extensions, extension)
			}
		}
	}
	return extensions
}
