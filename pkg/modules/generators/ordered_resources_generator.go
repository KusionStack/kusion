package generators

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/modules"
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
func NewOrderedResourcesGenerator(multipleOrderedKinds ...[]string) (modules.Generator, error) {
	orderedKinds := defaultOrderedKinds
	if len(multipleOrderedKinds) > 0 && len(multipleOrderedKinds[0]) > 0 {
		orderedKinds = multipleOrderedKinds[0]
	}
	return &orderedResourcesGenerator{
		orderedKinds: orderedKinds,
	}, nil
}

// NewOrderedResourcesGeneratorFunc returns a function that creates a new orderedResourcesGenerator.
func NewOrderedResourcesGeneratorFunc(multipleOrderedKinds ...[]string) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewOrderedResourcesGenerator(multipleOrderedKinds...)
	}
}

// Generate inject the dependsOn of resources in a specified order.
func (g *orderedResourcesGenerator) Generate(itt *v1.Spec) error {
	if itt.Resources == nil {
		itt.Resources = make(v1.Resources, 0)
	}

	for i := 0; i < len(itt.Resources); i++ {
		// Continue if the resource is not a kubernetes resource.
		if itt.Resources[i].Type != runtime.Kubernetes {
			continue
		}

		// Inject dependsOn of the resource.
		r := (*resource)(&itt.Resources[i])
		r.injectDependsOn(g.orderedKinds, itt.Resources)
	}

	return nil
}

type resource v1.Resource

// kubernetesKind returns the kubernetes kind of the given resource.
func (r resource) kubernetesKind() string {
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(r.Attributes)
	return u.GetKind()
}

// injectDependsOn injects all dependsOn relationships for the given resource and dependent kinds.
func (r *resource) injectDependsOn(orderedKinds []string, rs []v1.Resource) {
	kinds := r.findDependKinds(orderedKinds)
	for _, kind := range kinds {
		drs := findDependResources(kind, rs)
		r.appendDependsOn(drs)
	}
}

// appendDependsOn injects dependsOn relationships for the given resource and dependent resources.
func (r *resource) appendDependsOn(dependResources []*v1.Resource) {
	for _, dr := range dependResources {
		r.DependsOn = append(r.DependsOn, dr.ID)
	}
}

// findDependKinds returns the dependent resource kinds for the specified kind.
func (r *resource) findDependKinds(orderedKinds []string) []string {
	curKind := r.kubernetesKind()
	dependKinds := make([]string, 0)
	for _, previousKind := range orderedKinds {
		if curKind == previousKind {
			break
		}
		dependKinds = append(dependKinds, previousKind)
	}
	return dependKinds
}

// findDependResources returns the dependent resources of the specified kind.
func findDependResources(dependKind string, rs []v1.Resource) []*v1.Resource {
	var dependResources []*v1.Resource
	for i := 0; i < len(rs); i++ {
		if resource(rs[i]).kubernetesKind() == dependKind {
			dependResources = append(dependResources, &rs[i])
		}
	}
	return dependResources
}
