package generators

import (
	"context"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload/network"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/proto"
	jsonutil "kusionstack.io/kusion/pkg/util/json"

	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
)

func TestAppConfigurationGenerator_Generate(t *testing.T) {
	appName, app := buildMockApp()
	ws := buildMockWorkspace("")

	g := &appConfigurationGenerator{
		project: "fakeNs",
		stack:   "test",
		appName: appName,
		app:     app,
		ws:      ws,
	}

	spec := &v1.Intent{
		Resources: []v1.Resource{},
	}

	mockPlugin()
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
			ns := actual.GetNamespace()
			if ns == "" {
				// Manually get the namespace from the unstructed object.
				metadata, ok := actual.Object["metadata"].(map[interface{}]interface{})
				if !ok {
					t.Fatalf("failed to get metadata from unstructed object")
				}
				ns, ok = metadata["namespace"].(string)
				if !ok {
					t.Fatalf("failed to get namespace from metadata")
				}
			}
			assert.Equal(t, "fakeNs", ns, "namespace name should be fakeNs")
		}
	}
}

type fakeModule struct{}

func (f *fakeModule) Generate(ctx context.Context, req *proto.GeneratorRequest) (*proto.GeneratorResponse, error) {
	res := v1.Resource{
		ID:   "apps.kusionstack.io/v1alpha1:PodTransitionRule:fakeNs:default-dev-foo",
		Type: "Kubernetes",
		Attributes: map[string]interface{}{
			"apiVersion": "apps.kusionstack.io/v1alpha1",
			"kind":       "PodTransitionRule",
			"metadata": map[string]interface{}{
				"creationTimestamp": interface{}(nil),
				"name":              "default-dev-foo",
				"namespace":         "fakeNs",
			},
			"spec": map[string]interface{}{
				"rules": []interface{}{map[string]interface{}{
					"availablePolicy": map[string]interface{}{
						"maxUnavailableValue": "30%",
					},
					"name": "maxUnavailable",
				}},
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name": "foo", "app.kubernetes.io/part-of": "default",
					},
				},
			}, "status": map[string]interface{}{},
		},
		DependsOn: []string(nil),
		Extensions: map[string]interface{}{
			"GVK": "apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule",
		},
	}
	str := jsonutil.Marshal2String(res)
	b := []byte(str)
	return &proto.GeneratorResponse{
		Resources: [][]byte{b},
	}, nil
}

func mockPlugin() {
	mockey.Mock(modules.NewPlugin).To(func(key string) (*modules.Plugin, error) {
		return &modules.Plugin{Module: &fakeModule{}}, nil
	}).Build()
	mockey.Mock((*modules.Plugin).KillPluginClient).To(func() {
	}).Build()
}

func TestAppConfigurationGenerator_Generate_CustomNamespace(t *testing.T) {
	appName, app := buildMockApp()
	ws := buildMockWorkspace("fakeNs")

	g := &appConfigurationGenerator{
		project: "testproject",
		stack:   "test",
		appName: appName,
		app:     app,
		ws:      ws,
	}

	spec := &v1.Intent{
		Resources: []v1.Resource{},
	}

	mockPlugin()
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
			ns := actual.GetNamespace()
			if ns == "" {
				// Manually get the namespace from the unstructed object.
				metadata, ok := actual.Object["metadata"].(map[interface{}]interface{})
				if !ok {
					t.Fatalf("failed to get metadata from unstructed object")
				}
				ns, ok = metadata["namespace"].(string)
				if !ok {
					t.Fatalf("failed to get namespace from metadata")
				}
			}
			assert.Equal(t, "fakeNs", ns, "namespace name should be fakeNs")
		}
	}
}

func TestNewAppConfigurationGeneratorFunc(t *testing.T) {
	appName, app := buildMockApp()
	ws := buildMockWorkspace("")

	t.Run("Valid app configuration generator func", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", appName, app, ws)()
		assert.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Empty app name", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", "", app, ws)()
		assert.EqualError(t, err, "app name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Nil app", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", appName, nil, ws)()
		assert.EqualError(t, err, "can not find app configuration when generating the Intent")
		assert.Nil(t, g)
	})

	t.Run("Empty project name", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("", "test", appName, app, ws)()
		assert.EqualError(t, err, "project name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Empty workspace", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", appName, app, nil)()
		assert.EqualError(t, err, "workspace must not be empty")
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
			"kusionstack/database@v0.1": {
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

func mapToUnstructured(data map[string]interface{}) *unstructured.Unstructured {
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(data)
	return unstructuredObj
}
