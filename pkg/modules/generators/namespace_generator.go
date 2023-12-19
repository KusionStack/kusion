package generators

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/apis/intent"
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/workspace"
)

type namespaceGenerator struct {
	projectName  string
	moduleInputs map[string]workspaceapi.GenericConfig
}

func NewNamespaceGenerator(projectName string, ws *workspaceapi.Workspace) (modules.Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}
	moduleInputs, err := workspace.GetProjectModuleConfigs(ws.Modules, projectName)
	if err != nil {
		return nil, fmt.Errorf("parse project matched module configs failed: %v", err)
	}

	return &namespaceGenerator{
		projectName:  projectName,
		moduleInputs: moduleInputs,
	}, nil
}

func NewNamespaceGeneratorFunc(projectName string, workspace *workspaceapi.Workspace) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewNamespaceGenerator(projectName, workspace)
	}
}

func (g *namespaceGenerator) Generate(i *intent.Intent) error {
	if i.Resources == nil {
		i.Resources = make(intent.Resources, 0)
	}

	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: g.getName(g.projectName)},
	}

	// Avoid generating duplicate namespaces with the same ID.
	id := modules.KubernetesResourceID(ns.TypeMeta, ns.ObjectMeta)
	for _, res := range i.Resources {
		if res.ID == id {
			return nil
		}
	}

	return modules.AppendToIntent(intent.Kubernetes, id, i, ns)
}

// getName obtains the name for this Namespace using the following precedence
// (from lower to higher):
// - Project name
// - Namespace module config (specified in corresponding workspace file)
func (g *namespaceGenerator) getName(projectName string) string {
	if g.moduleInputs == nil {
		return projectName
	}

	namespaceName := projectName
	namespaceModuleConfigs, exist := g.moduleInputs["namespace"]
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
