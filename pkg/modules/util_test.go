package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
)

type mockGenerator struct {
	GenerateFunc func(intent *intent.Intent) error
}

func (m *mockGenerator) Generate(i *intent.Intent) error {
	return m.GenerateFunc(i)
}

type mockPatcher struct {
	PatchFunc func(resources map[string][]*intent.Resource) error
}

func (m *mockPatcher) Patch(resources map[string][]*intent.Resource) error {
	return m.PatchFunc(resources)
}

func TestCallGenerators(t *testing.T) {
	i := &intent.Intent{}

	var (
		generator1 Generator = &mockGenerator{
			GenerateFunc: func(intent *intent.Intent) error {
				return nil
			},
		}
		generator2 Generator = &mockGenerator{
			GenerateFunc: func(intent *intent.Intent) error {
				return assert.AnError
			},
		}
		gf1 = func() (Generator, error) { return generator1, nil }
		gf2 = func() (Generator, error) { return generator2, nil }
	)

	err := CallGenerators(i, gf1, gf2)
	assert.Error(t, err)
	assert.EqualError(t, err, assert.AnError.Error())
}

func TestCallPatchers(t *testing.T) {
	var (
		patcher1 Patcher = &mockPatcher{
			PatchFunc: func(resources map[string][]*intent.Resource) error {
				return nil
			},
		}
		patcher2 Patcher = &mockPatcher{
			PatchFunc: func(resources map[string][]*intent.Resource) error {
				return assert.AnError
			},
		}
		pf1 = func() (Patcher, error) { return patcher1, nil }
		pf2 = func() (Patcher, error) { return patcher2, nil }
	)
	err := CallPatchers(nil, pf1)
	assert.NoError(t, err)

	err = CallPatchers(nil, pf1, pf2)
	assert.Error(t, err)
	assert.EqualError(t, err, assert.AnError.Error())
}

func TestCallGeneratorFuncs(t *testing.T) {
	generatorFunc1 := func() (Generator, error) {
		return &mockGenerator{}, nil
	}

	generatorFunc2 := func() (Generator, error) {
		return nil, assert.AnError
	}

	generators, err := CallGeneratorFuncs(generatorFunc1)
	assert.NoError(t, err)
	assert.Len(t, generators, 1)
	assert.IsType(t, &mockGenerator{}, generators[0])

	_, err = CallGeneratorFuncs(generatorFunc2)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestForeachOrdered(t *testing.T) {
	m := map[string]int{
		"b": 2,
		"a": 1,
		"c": 3,
	}

	result := ""
	err := ForeachOrdered(m, func(key string, value int) error {
		result += key
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "abc", result)
}

func TestGenericPtr(t *testing.T) {
	value := 42
	ptr := GenericPtr(value)
	assert.Equal(t, &value, ptr)
}

func TestMergeMaps(t *testing.T) {
	map1 := map[string]string{
		"a": "1",
		"b": "2",
	}

	map2 := map[string]string{
		"c": "3",
		"d": "4",
	}

	merged := MergeMaps(map1, nil, map2)

	expected := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
	}

	assert.Equal(t, expected, merged)
}

func TestKubernetesResourceID(t *testing.T) {
	typeMeta := v1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}

	objectMeta := v1.ObjectMeta{
		Namespace: "example",
		Name:      "my-deployment",
	}

	id := KubernetesResourceID(typeMeta, objectMeta)
	assert.Equal(t, "apps/v1:Deployment:example:my-deployment", id)
}

func TestAppendToIntent(t *testing.T) {
	i := &intent.Intent{}
	resource := &intent.Resource{
		ID:   "v1:Namespace:fake-project",
		Type: "Kubernetes",
		Attributes: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"name":              "fake-project",
			},
			"spec":   make(map[string]interface{}),
			"status": make(map[string]interface{}),
		},
		DependsOn:  nil,
		Extensions: nil,
	}

	ns := &corev1.Namespace{
		TypeMeta: v1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "fake-project",
		},
	}

	err := AppendToIntent(intent.Kubernetes, resource.ID, i, ns)

	assert.NoError(t, err)
	assert.Len(t, i.Resources, 1)
	assert.Equal(t, resource.ID, i.Resources[0].ID)
	assert.Equal(t, resource.Type, i.Resources[0].Type)
	assert.Equal(t, resource.Attributes, i.Resources[0].Attributes)
	assert.Equal(t, ns.GroupVersionKind().String(), i.Resources[0].Extensions[intent.ResourceExtensionGVK])
}

func TestUniqueAppName(t *testing.T) {
	projectName := "my-project"
	stackName := "my-stack"
	appName := "my-app"

	expected := "my-project-my-stack-my-app"
	result := UniqueAppName(projectName, stackName, appName)

	assert.Equal(t, expected, result)
}

func TestUniqueAppLabels(t *testing.T) {
	projectName := "my-project"
	appName := "my-app"

	expected := map[string]string{
		"app.kubernetes.io/part-of": projectName,
		"app.kubernetes.io/name":    appName,
	}

	result := UniqueAppLabels(projectName, appName)

	assert.Equal(t, expected, result)
}

func TestPatchResource(t *testing.T) {
	resources := map[string][]*intent.Resource{
		"/v1, Kind=Namespace": {
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "default",
					},
				},
				Extensions: map[string]interface{}{
					"GVK": "/v1, Kind=Namespace",
				},
			},
		},
	}
	assert.NoError(
		t,
		PatchResource(resources, "/v1, Kind=Namespace", func(ns *corev1.Namespace) error {
			ns.Labels = map[string]string{
				"foo": "bar",
			}
			return nil
		}),
	)
	assert.Equal(
		t,
		map[string]interface{}{
			"foo": "bar",
		},
		resources["/v1, Kind=Namespace"][0].Attributes["metadata"].(map[string]interface{})["labels"].(map[string]interface{}),
	)
}

func TestAddKubeConfigIf(t *testing.T) {
	testcases := []struct {
		name           string
		ws             *apiv1.Workspace
		i              *intent.Intent
		expectedIntent *intent.Intent
	}{
		{
			name: "empty workspace runtime config",
			ws:   &apiv1.Workspace{Name: "dev"},
			i: &intent.Intent{
				Resources: intent.Resources{
					{
						ID:   "mock-id-1",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
				},
			},
			expectedIntent: &intent.Intent{
				Resources: intent.Resources{
					{
						ID:   "mock-id-1",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
				},
			},
		},
		{
			name: "empty kubeConfig in workspace",
			ws: &apiv1.Workspace{
				Name: "dev",
				Runtimes: &apiv1.RuntimeConfigs{
					Kubernetes: &apiv1.KubernetesConfig{},
				},
			},
			i: &intent.Intent{
				Resources: intent.Resources{
					{
						ID:   "mock-id-1",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
				},
			},
			expectedIntent: &intent.Intent{
				Resources: intent.Resources{
					{
						ID:   "mock-id-1",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
				},
			},
		},
		{
			name: "add kubeConfig",
			ws: &apiv1.Workspace{
				Name: "dev",
				Runtimes: &apiv1.RuntimeConfigs{
					Kubernetes: &apiv1.KubernetesConfig{
						KubeConfig: "/etc/kubeConfig.yaml",
					},
				},
			},
			i: &intent.Intent{
				Resources: intent.Resources{
					{
						ID:   "mock-id-1",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
					{
						ID:   "mock-id-2",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"mock-extensions-key": "mock-extensions-value",
						},
					},
					{
						ID:   "mock-id-2",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"kubeConfig": "/etc/should-use-kubeConfig.yaml",
						},
					},
					{
						ID:   "mock-id-3",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"kubeConfig": "",
						},
					},
					{
						ID:   "mock-id-4",
						Type: "Terraform",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
				},
			},
			expectedIntent: &intent.Intent{
				Resources: intent.Resources{
					{
						ID:   "mock-id-1",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"kubeConfig": "/etc/kubeConfig.yaml",
						},
					},
					{
						ID:   "mock-id-2",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"mock-extensions-key": "mock-extensions-value",
							"kubeConfig":          "/etc/kubeConfig.yaml",
						},
					},
					{
						ID:   "mock-id-2",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"kubeConfig": "/etc/should-use-kubeConfig.yaml",
						},
					},
					{
						ID:   "mock-id-3",
						Type: "Kubernetes",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: map[string]any{
							"kubeConfig": "/etc/kubeConfig.yaml",
						},
					},
					{
						ID:   "mock-id-4",
						Type: "Terraform",
						Attributes: map[string]any{
							"mock-key": "mock-value",
						},
						Extensions: nil,
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			AddKubeConfigIf(tc.i, tc.ws)
			assert.Equal(t, *tc.expectedIntent, *tc.i)
		})
	}
}
