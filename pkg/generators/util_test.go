package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

type mockGenerator struct {
	GenerateFunc func(Spec *v1.Spec) error
}

func (m *mockGenerator) Generate(i *v1.Spec) error {
	return m.GenerateFunc(i)
}

func TestCallGenerators(t *testing.T) {
	i := &v1.Spec{}

	var (
		generator1 SpecGenerator = &mockGenerator{
			GenerateFunc: func(Spec *v1.Spec) error {
				return nil
			},
		}
		generator2 SpecGenerator = &mockGenerator{
			GenerateFunc: func(Spec *v1.Spec) error {
				return assert.AnError
			},
		}
		gf1 = func() (SpecGenerator, error) { return generator1, nil }
		gf2 = func() (SpecGenerator, error) { return generator2, nil }
	)

	err := CallGenerators(i, gf1, gf2)
	assert.Error(t, err)
	assert.EqualError(t, err, assert.AnError.Error())
}

func TestCallGeneratorFuncs(t *testing.T) {
	generatorFunc1 := func() (SpecGenerator, error) {
		return &mockGenerator{}, nil
	}

	generatorFunc2 := func() (SpecGenerator, error) {
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

func TestAppendToSpec(t *testing.T) {
	i := &v1.Spec{}
	resource := &v1.Resource{
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
	}

	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-project",
		},
	}

	err := AppendToSpec(v1.Kubernetes, resource.ID, i, ns)

	assert.NoError(t, err)
	assert.Len(t, i.Resources, 1)
	assert.Equal(t, resource.ID, i.Resources[0].ID)
	assert.Equal(t, resource.Type, i.Resources[0].Type)
	assert.Equal(t, resource.Attributes, i.Resources[0].Attributes)
	assert.Equal(t, ns.GroupVersionKind().String(), i.Resources[0].Extensions[v1.ResourceExtensionGVK])
}

func TestAddKubeConfigIf(t *testing.T) {
	testcases := []struct {
		name         string
		ws           *v1.Workspace
		i            *v1.Spec
		expectedSpec *v1.Spec
	}{
		{
			name: "empty workspace runtime config",
			ws:   &v1.Workspace{Name: "dev"},
			i: &v1.Spec{
				Resources: v1.Resources{
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
			expectedSpec: &v1.Spec{
				Resources: v1.Resources{
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
			ws: &v1.Workspace{
				Name: "dev",
			},
			i: &v1.Spec{
				Resources: v1.Resources{
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
			expectedSpec: &v1.Spec{
				Resources: v1.Resources{
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, *tc.expectedSpec, *tc.i)
		})
	}
}
