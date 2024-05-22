package modules

import (
	"errors"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// IgnoreModules todo@dayuan delete this condition after workload is changed into a module
var IgnoreModules = map[string]bool{
	"service": true,
	"job":     true,
}

// CallGeneratorFuncs calls each NewGeneratorFunc in the given slice
// and returns a slice of Generator instances.
func CallGeneratorFuncs(newGenerators ...NewGeneratorFunc) ([]Generator, error) {
	gs := make([]Generator, 0, len(newGenerators))
	for _, newGenerator := range newGenerators {
		if g, err := newGenerator(); err != nil {
			return nil, err
		} else {
			gs = append(gs, g)
		}
	}
	return gs, nil
}

// CallGenerators calls the Generate method of each Generator instance
// returned by the given NewGeneratorFuncs.
func CallGenerators(i *v1.Spec, newGenerators ...NewGeneratorFunc) error {
	gs, err := CallGeneratorFuncs(newGenerators...)
	if err != nil {
		return err
	}
	for _, g := range gs {
		if err := g.Generate(i); err != nil {
			return err
		}
	}
	return nil
}

// ForeachOrdered executes the given function on each
// item in the map in order of their keys.
func ForeachOrdered[T any](m map[string]T, f func(key string, value T) error) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}

// GenericPtr returns a pointer to the provided value.
func GenericPtr[T any](i T) *T {
	return &i
}

// MergeMaps merges multiple map[string]string into one
// map[string]string.
// If a map is nil, it skips it and moves on to the next one. For each
// non-nil map, it iterates over its key-value pairs and adds them to
// the merged map. Finally, it returns the merged map.
func MergeMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)

	for _, m := range maps {
		if len(m) == 0 {
			continue
		}
		for k, v := range m {
			merged[k] = v
		}
	}

	if len(merged) == 0 {
		return nil
	}
	return merged
}

// KubernetesResourceID returns the unique ID of a Kubernetes resource
// based on its type and metadata.
func KubernetesResourceID(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:code-city:code-citydev
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id += objectMeta.Namespace + ":"
	}
	id += objectMeta.Name
	return id
}

// KusionPathDependency returns the implicit resource dependency path based on
// the resource id and name with the "$kusion_path" prefix.
func KusionPathDependency(id, name string) string {
	return "$kusion_path." + id + "." + name
}

// AppendToSpec adds a Kubernetes resource to the Spec resources slice.
func AppendToSpec(resourceType v1.Type, resourceID string, i *v1.Spec, resource any) error {
	// this function is only used for Kubernetes resources
	if resourceType != v1.Kubernetes {
		return errors.New("AppendToSpec is only used for Kubernetes resources")
	}

	gvk := resource.(runtime.Object).GetObjectKind().GroupVersionKind().String()
	// fixme: this function converts int to int64 by default
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	r := v1.Resource{
		ID:         resourceID,
		Type:       resourceType,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: map[string]any{
			v1.ResourceExtensionGVK: gvk,
		},
	}
	i.Resources = append(i.Resources, r)
	return nil
}

// UniqueAppName returns a unique name for a workload based on its project and app name.
func UniqueAppName(projectName, stackName, appName string) string {
	return projectName + "-" + stackName + "-" + appName
}

// UniqueAppLabels returns a map of labels that identify an app based on its project and name.
func UniqueAppLabels(projectName, appName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of": projectName,
		"app.kubernetes.io/name":    appName,
	}
}

// PatchResource patches the resource with the given patch.
func PatchResource[T any](resources map[string][]*v1.Resource, gvk string, patchFunc func(*T) error) error {
	var obj T
	for _, r := range resources[gvk] {
		// convert unstructured to typed object
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.Attributes, &obj); err != nil {
			return err
		}

		if err := patchFunc(&obj); err != nil {
			return err
		}

		// convert typed object to unstructured
		updated, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
		if err != nil {
			return err
		}
		r.Attributes = updated
	}
	return nil
}
