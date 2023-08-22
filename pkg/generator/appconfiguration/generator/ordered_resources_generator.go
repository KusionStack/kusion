package generator

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
)

// defaultOrderedKinds provides the default order of kubernetes resource kinds.
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

// orderedResourcesGenerator is a generator that inject the dependsOn of resources in a specified order.
type orderedResourcesGenerator struct {
	orderedKinds []string
}

// NewOrderedResourcesGenerator returns a new instance of orderedResourcesGenerator.
func NewOrderedResourcesGenerator(multipleOrderedKinds ...[]string) (appconfiguration.Generator, error) {
	orderedKinds := defaultOrderedKinds
	if len(multipleOrderedKinds) > 0 && len(multipleOrderedKinds[0]) > 0 {
		orderedKinds = multipleOrderedKinds[0]
	}
	return &orderedResourcesGenerator{
		orderedKinds: orderedKinds,
	}, nil
}

// NewOrderedResourcesGeneratorFunc returns a function that creates a new orderedResourcesGenerator.
func NewOrderedResourcesGeneratorFunc(multipleOrderedKinds ...[]string) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewOrderedResourcesGenerator(multipleOrderedKinds...)
	}
}

// Generate inject the dependsOn of resources in a specified order.
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

// resourceKind returns the kind of the given resource.
func resourceKind(r *models.Resource) string {
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(r.Attributes)
	return u.GetKind()
}

// injectAllDependsOn injects all dependsOn relationships for the given resource and dependent kinds.
func injectAllDependsOn(curResource *models.Resource, dependKinds []string, rs []models.Resource) {
	for _, dependKind := range dependKinds {
		dependResources := findDependResources(dependKind, rs)
		injectDependsOn(curResource, dependResources)
	}
}

// injectDependsOn injects dependsOn relationships for the given resource and dependent resources.
func injectDependsOn(res *models.Resource, dependResources []*models.Resource) {
	dependsOn := make([]string, 0, len(dependResources))
	for _, r := range dependResources {
		res.DependsOn = append(res.DependsOn, r.ID)
	}
	if len(dependsOn) > 0 {
		res.DependsOn = dependsOn
	}
}

// findDependResources returns the dependent resources of the specified kind.
func findDependResources(dependKind string, rs []models.Resource) []*models.Resource {
	var dependResources []*models.Resource
	for i := 0; i < len(rs); i++ {
		if resourceKind(&rs[i]) == dependKind {
			dependResources = append(dependResources, &rs[i])
		}
	}
	return dependResources
}

// findDependKinds returns the dependent resource kinds for the specified kind.
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
