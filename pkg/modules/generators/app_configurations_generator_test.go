package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload/network"

	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
)

func TestAppConfigurationGenerator_Generate(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()
	ws := buildMockWorkspace("")

	g := &appConfigurationGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
		ws:      ws,
	}

	spec := &v1.Intent{
		Resources: []v1.Resource{},
	}

	err := g.Generate(spec)
	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Resources)

	// namespace name assertion
	for _, res := range spec.Resources {
		if res.Type != v1.Kubernetes {
			continue
		}
		actual := mapToUnstructured(res.Attributes)
		if actual.GetKind() == "Namespace" {
			assert.Equal(t, "testproject", actual.GetName(), "namespace name should be fakeNs")
		} else {
			assert.Equal(t, "testproject", actual.GetNamespace(), "namespace name should be fakeNs")
		}
	}
}

func TestAppConfigurationGenerator_Generate_CustomNamespace(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()
	ws := buildMockWorkspace("fakeNs")

	g := &appConfigurationGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
		ws:      ws,
	}

	spec := &v1.Intent{
		Resources: []v1.Resource{},
	}

	err := g.Generate(spec)
	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Resources)

	// namespace name assertion
	for _, res := range spec.Resources {
		if res.Type != v1.Kubernetes {
			continue
		}
		actual := mapToUnstructured(res.Attributes)
		if actual.GetKind() == "Namespace" {
			assert.Equal(t, "fakeNs", actual.GetName(), "namespace name should be fakeNs")
		} else {
			assert.Equal(t, "fakeNs", actual.GetNamespace(), "namespace name should be fakeNs")
		}
	}
}

func TestNewAppConfigurationGeneratorFunc(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()
	ws := buildMockWorkspace("")

	t.Run("Valid app configuration generator func", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, app, ws)()
		assert.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Empty app name", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, "", app, ws)()
		assert.EqualError(t, err, "app name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Nil app", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, nil, ws)()
		assert.EqualError(t, err, "can not find app configuration when generating the Intent")
		assert.Nil(t, g)
	})

	t.Run("Empty project name", func(t *testing.T) {
		project.Name = ""
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, app, ws)()
		assert.EqualError(t, err, "project name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Empty workspace", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, app, nil)()
		assert.EqualError(t, err, "project name must not be empty")
		assert.Nil(t, g)
	})
}

func buildMockApp() (string, *v1.AppConfiguration) {
	return "app1", &v1.AppConfiguration{
		Workload: &workload.Workload{
			Header: workload.Header{
				Type: "Service",
			},
			Service: &workload.Service{
				Base: workload.Base{},
				Type: "Deployment",
				Ports: []network.Port{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
			},
		},
	}
}

func buildMockWorkspace(namespace string) *v1.Workspace {
	return &v1.Workspace{
		Name: "test",
		Modules: v1.ModuleConfigs{
			"database": {
				Default: v1.GenericConfig{
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.micro",
				},
				ModulePatcherConfigs: v1.ModulePatcherConfigs{
					"smallClass": {
						GenericConfig: v1.GenericConfig{
							"instanceType": "db.t3.small",
						},
						ProjectSelector: []string{"foo", "bar"},
					},
				},
			},
			"port": {
				Default: v1.GenericConfig{
					"type": "aws",
				},
			},
			"namespace": {
				Default: v1.GenericConfig{
					"name": namespace,
				},
			},
		},
		Runtimes: &v1.RuntimeConfigs{
			Kubernetes: &v1.KubernetesConfig{
				KubeConfig: "/etc/kubeconfig.yaml",
			},
		},
		Backends: &v1.DeprecatedBackendConfigs{
			Local: &v1.DeprecatedLocalFileConfig{},
		},
	}
}

func buildMockProjectAndStack() (*v1.Project, *v1.Stack) {
	project := &v1.Project{
		Name: "testproject",
	}

	stack := &v1.Stack{
		Name: "test",
	}

	return project, stack
}

func mapToUnstructured(data map[string]interface{}) *unstructured.Unstructured {
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(data)
	return unstructuredObj
}
