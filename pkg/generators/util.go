package generators

import (
	"errors"
	"sort"

	"k8s.io/apimachinery/pkg/runtime"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// CallGeneratorFuncs calls each NewSpecGeneratorFunc in the given slice
// and returns a slice of SpecGenerator instances.
func CallGeneratorFuncs(newGenerators ...NewSpecGeneratorFunc) ([]SpecGenerator, error) {
	gs := make([]SpecGenerator, 0, len(newGenerators))
	for _, newGenerator := range newGenerators {
		if g, err := newGenerator(); err != nil {
			return nil, err
		} else {
			gs = append(gs, g)
		}
	}
	return gs, nil
}

// CallGenerators calls the Generate method of each SpecGenerator instance
// returned by the given NewSpecGeneratorFuncs.
func CallGenerators(i *v1.Spec, newGenerators ...NewSpecGeneratorFunc) error {
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
