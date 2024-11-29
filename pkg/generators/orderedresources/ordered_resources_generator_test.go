package orderedresources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/runtime"
)

var (
	fakeDeployment = map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"namespace": "foo",
			"name":      "bar",
		},
		"spec": map[string]interface{}{
			"replica": 1,
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"image": "foo.bar.com:v1",
							"name":  "bar",
						},
					},
				},
			},
		},
	}
	fakeService = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"namespace": "foo",
			"name":      "bar",
		},
	}
	fakeNamespace = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": "foo",
		},
	}
	genOldSpec = func() *v1.Spec {
		return &v1.Spec{
			Resources: v1.Resources{
				{
					ID:         "apps/v1:Deployment:foo:bar",
					Type:       runtime.Kubernetes,
					Attributes: fakeDeployment,
				},
				{
					ID:         "v1:Service:foo:bar",
					Type:       runtime.Kubernetes,
					Attributes: fakeService,
				},
				{
					ID:         "v1:Namespace:foo",
					Type:       runtime.Kubernetes,
					Attributes: fakeNamespace,
				},
			},
		}
	}
	genNewSpec = func() *v1.Spec {
		return &v1.Spec{
			Resources: v1.Resources{
				{
					ID:         "apps/v1:Deployment:foo:bar",
					Type:       runtime.Kubernetes,
					Attributes: fakeDeployment,
					DependsOn:  []string{"v1:Namespace:foo", "v1:Service:foo:bar"},
				},
				{
					ID:         "v1:Service:foo:bar",
					Type:       runtime.Kubernetes,
					Attributes: fakeService,
					DependsOn:  []string{"v1:Namespace:foo"},
				},
				{
					ID:         "v1:Namespace:foo",
					Type:       runtime.Kubernetes,
					Attributes: fakeNamespace,
				},
			},
		}
	}
)

func TestOrderedResourcesGenerator_Generate(t *testing.T) {
	orderedGenerator, err := NewOrderedResourcesGenerator()
	assert.NoError(t, err)

	expected := genNewSpec()
	actual := genOldSpec()
	err = orderedGenerator.Generate(actual)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
