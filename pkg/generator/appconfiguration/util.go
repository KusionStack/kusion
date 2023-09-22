package appconfiguration

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kusion/pkg/models"
)

// CallGeneratorFuncs calls each NewGeneratorFunc in the given slice
// and returns a slice of AppsGenerator instances.
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

// CallGenerators calls the Generate method of each AppsGenerator instance
// returned by the given NewGeneratorFuncs.
func CallGenerators(spec *models.Spec, newGenerators ...NewGeneratorFunc) error {
	gs, err := CallGeneratorFuncs(newGenerators...)
	if err != nil {
		return err
	}
	for _, g := range gs {
		if err := g.Generate(spec); err != nil {
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

// TerraformResource returns the Terraform resource in the form of
// Kusion's spec resource.
func TerraformResource(id string, dependsOn []string, attrs, exts map[string]interface{}) models.Resource {
	return models.Resource{
		ID:         id,
		Type:       models.Terraform,
		Attributes: attrs,
		DependsOn:  dependsOn,
		Extensions: exts,
	}
}

// TerraformResourceID returns the unique ID of a Terraform resource
// based on its provider, type and name.
func TerraformResourceID(provider *models.Provider, resourceType string, resourceName string) string {
	// resource id example: hashicorp:aws:aws_db_instance:wordpressdev
	return provider.Namespace + ":" + provider.Name + ":" + resourceType + ":" + resourceName
}

// ProviderExtensions returns the extended information of provider based on
// the provider and type of the resource.
func ProviderExtensions(provider *models.Provider, providerMeta map[string]any, resourceType string) map[string]interface{} {
	return map[string]interface{}{
		"provider":     provider.URL,
		"providerMeta": providerMeta,
		"resourceType": resourceType,
	}
}

// KusionPathDependency returns the implicit resource dependency path based on
// the resource id and name with the "$kusion_path" prefix.
func KusionPathDependency(id, name string) string {
	return "$kusion_path." + id + "." + name
}

// AppendToSpec adds a Kubernetes resource to a spec's resources slice.
func AppendToSpec(resourceType models.Type, resourceID string, spec *models.Spec, resource any) error {
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	r := models.Resource{
		ID:         resourceID,
		Type:       resourceType,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: nil,
	}
	spec.Resources = append(spec.Resources, r)
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
