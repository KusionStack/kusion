package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/models"
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
	fakeNamespace = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": "foo",
		},
	}
	genOldSpec = func() *models.Spec {
		return &models.Spec{
			Resources: models.Resources{
				{
					ID:         "apps/v1:Deployment:foo:bar",
					Type:       runtime.Kubernetes,
					Attributes: fakeDeployment,
				},
				{
					ID:         "v1:Namespace:foo",
					Type:       runtime.Kubernetes,
					Attributes: fakeNamespace,
				},
			},
		}
	}
	genNewSpec = func() *models.Spec {
		return &models.Spec{
			Resources: models.Resources{
				{
					ID:         "apps/v1:Deployment:foo:bar",
					Type:       runtime.Kubernetes,
					Attributes: fakeDeployment,
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
	resource := &models.Resource{
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"kind": "Deployment",
		},
	}

	expected := "Deployment"
	actual := resourceKind(resource)

	assert.Equal(t, expected, actual)
}

func TestInjectAllDependsOn(t *testing.T) {
	spec := genOldSpec()
	dependKinds := []string{"Namespace"}

	expected := []string{"v1:Namespace:foo"}
	actual := []models.Resource(spec.Resources)[0]
	injectAllDependsOn(&actual, dependKinds, spec.Resources)

	assert.Equal(t, expected, actual.DependsOn)
}

func TestFindDependKinds(t *testing.T) {
	curKind := "Deployment"
	g := &orderedResourcesGenerator{
		orderedKinds: defaultOrderedKinds,
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
	actual := g.findDependKinds(curKind)

	assert.Equal(t, expected, actual)
}

func TestFindDependResources(t *testing.T) {
	dependKind := "Namespace"
	resources := genOldSpec().Resources

	expected := []*models.Resource{
		{
			ID:         "v1:Namespace:foo",
			Type:       runtime.Kubernetes,
			Attributes: fakeNamespace,
		},
	}
	actual := findDependResources(dependKind, resources)

	assert.Equal(t, expected, actual)
}
