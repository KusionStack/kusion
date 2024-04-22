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

package builders

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	internalv1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators"
)

type AppsConfigBuilder struct {
	Apps      map[string]internalv1.AppConfiguration
	Workspace *v1.Workspace
}

func (acg *AppsConfigBuilder) Build(kclPackage *api.KclPackage, project *v1.Project, stack *v1.Stack) (*v1.Spec, error) {
	i := &v1.Spec{
		Resources: []v1.Resource{},
	}

	var gfs []modules.NewGeneratorFunc
	err := modules.ForeachOrdered(acg.Apps, func(appName string, app internalv1.AppConfiguration) error {
		if kclPackage == nil {
			return fmt.Errorf("kcl package is nil when generating app configuration for %s", appName)
		}
		dependencies := kclPackage.GetDependenciesInModFile()
		gfs = append(gfs, generators.NewAppConfigurationGeneratorFunc(project.Name, stack.Name, appName, &app, acg.Workspace, dependencies))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = modules.CallGenerators(i, gfs...); err != nil {
		return nil, err
	}

	// updates generated spec resources based on project and stack extensions.
	patchResourcesWithExtensions(project, stack, i)

	return i, nil
}

// patchResourcesWithExtensions updates generated spec resources based on project and stack extensions.
func patchResourcesWithExtensions(project *v1.Project, stack *v1.Stack, spec *v1.Spec) {
	extensions := mergeExtensions(project, stack)
	if len(extensions) == 0 {
		return
	}

	for _, extension := range extensions {
		switch extension.Kind {
		case v1.KubernetesNamespace:
			patchResourcesKubeNamespace(spec, extension.KubeNamespace.Namespace)
		default:
			// do nothing
		}
	}
}

func patchResourcesKubeNamespace(spec *v1.Spec, namespace string) {
	for _, resource := range spec.Resources {
		if resource.Type == v1.Kubernetes {
			u := &unstructured.Unstructured{Object: resource.Attributes}
			u.SetNamespace(namespace)
		}
	}
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
