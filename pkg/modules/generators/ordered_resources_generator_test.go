package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/intent"
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
	genOldSpec = func() *intent.Intent {
		return &intent.Intent{
			Resources: intent.Resources{
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
	genNewSpec = func() *intent.Intent {
		return &intent.Intent{
			Resources: intent.Resources{
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

func TestResourceKind(t *testing.T) {
	r := &resource{
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"kind": "Deployment",
		},
	}

	expected := "Deployment"
	actual := r.kubernetesKind()

	assert.Equal(t, expected, actual)
}

func TestInjectAllDependsOn(t *testing.T) {
	spec := genOldSpec()
	dependKinds := []string{"Namespace"}

	expected := []string{"v1:Namespace:foo"}
	actual := resource([]intent.Resource(spec.Resources)[0])
	actual.injectDependsOn(dependKinds, spec.Resources)

	assert.Equal(t, expected, actual.DependsOn)
}

func TestFindDependKinds(t *testing.T) {
	r := &resource{
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"kind": "Deployment",
		},
	}

	expected := []string{
		"Namespace",
		"ResourceQuota",
		"StorageClass",
		"CustomResourceDefinition",
		"ServiceAccount",
		"PodSecurityPolicy",
		"Role",
		"ClusterRole",
		"RoleBinding",
		"ClusterRoleBinding",
		"ConfigMap",
		"Secret",
		"Endpoints",
		"Service",
		"LimitRange",
		"PriorityClass",
		"PersistentVolume",
		"PersistentVolumeClaim",
	}
	actual := r.findDependKinds(defaultOrderedKinds)

	assert.Equal(t, expected, actual)
}

func TestFindDependResources(t *testing.T) {
	dependKind := "Namespace"
	resources := genOldSpec().Resources

	expected := []*intent.Resource{
		{
			ID:         "v1:Namespace:foo",
			Type:       runtime.Kubernetes,
			Attributes: fakeNamespace,
		},
	}
	actual := findDependResources(dependKind, resources)

	assert.Equal(t, expected, actual)
}
