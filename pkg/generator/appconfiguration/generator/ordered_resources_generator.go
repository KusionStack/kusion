package generator

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
)

var defaultOrderedKinds = []string{
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
	"Deployment",
	"StatefulSet",
	"CronJob",
	"PodDisruptionBudget",
	"MutatingWebhookConfiguration",
	"ValidatingWebhookConfiguration",
}

type orderedResourcesGenerator struct {
	orderedKinds []string
}

func NewOrderedResourcesGenerator(orderedKindsList ...[]string) (appconfiguration.Generator, error) {
	orderedKinds := defaultOrderedKinds
	if len(orderedKindsList) > 0 && len(orderedKindsList[0]) > 0 {
		orderedKinds = orderedKindsList[0]
	}
	return &orderedResourcesGenerator{
		orderedKinds: orderedKinds,
	}, nil
}

func NewOrderedResourcesGeneratorFunc(orderedKindsList ...[]string) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewOrderedResourcesGenerator(orderedKindsList...)
	}
}

func (g *orderedResourcesGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	for i := 0; i < len(spec.Resources); i++ {
		if spec.Resources[i].Type != runtime.Kubernetes {
			continue
		}

		curResource := &spec.Resources[i]
		curKind := resourceKind(curResource)

		dependKinds := g.findDependKinds(curKind)
		injectAllDependsOn(curResource, dependKinds, spec.Resources)
	}

	return nil
}

func resourceKind(r *models.Resource) string {
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(r.Attributes)
	return u.GetKind()
}

func injectAllDependsOn(curResource *models.Resource, dependKinds []string, rs []models.Resource) {
	for _, dependKind := range dependKinds {
		dependResources := findDependResources(dependKind, rs)
		injectDependsOn(curResource, dependResources)
	}
}

func injectDependsOn(res *models.Resource, dependResources []*models.Resource) {
	dependsOn := make([]string, 0, len(dependResources))
	for _, r := range dependResources {
		res.DependsOn = append(res.DependsOn, r.ID)
	}
	if len(dependsOn) > 0 {
		res.DependsOn = dependsOn
	}
}

func findDependResources(dependKind string, rs []models.Resource) []*models.Resource {
	var dependResources []*models.Resource
	for i := 0; i < len(rs); i++ {
		if resourceKind(&rs[i]) == dependKind {
			dependResources = append(dependResources, &rs[i])
		}
	}
	return dependResources
}

func (g *orderedResourcesGenerator) findDependKinds(curKind string) []string {
	dependKinds := make([]string, 0)
	for _, previousKind := range g.orderedKinds {
		if curKind == previousKind {
			break
		}
		dependKinds = append(dependKinds, previousKind)
	}
	return dependKinds
}
