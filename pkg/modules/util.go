package modules

import (
	"errors"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/apis/intent"
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/workspace"
)

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
func CallGenerators(i *intent.Intent, newGenerators ...NewGeneratorFunc) error {
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

// CallPatchers calls the Patch method of each Generator instance
// returned by the given NewPatcherFuncs.
func CallPatchers(resources map[string][]*intent.Resource, newPatchers ...NewPatcherFunc) error {
	ps := make([]Patcher, 0, len(newPatchers))
	for _, newPatcher := range newPatchers {
		if p, err := newPatcher(); err != nil {
			return err
		} else {
			ps = append(ps, p)
		}
	}
	for _, p := range ps {
		if err := p.Patch(resources); err != nil {
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

// TerraformResource returns the Terraform resource in the form of Intent.Resource
func TerraformResource(id string, dependsOn []string, attrs, exts map[string]interface{}) intent.Resource {
	return intent.Resource{
		ID:         id,
		Type:       intent.Terraform,
		Attributes: attrs,
		DependsOn:  dependsOn,
		Extensions: exts,
	}
}

// TerraformResourceID returns the unique ID of a Terraform resource
// based on its provider, type and name.
func TerraformResourceID(provider *inputs.Provider, resourceType string, resourceName string) string {
	// resource id example: hashicorp:aws:aws_db_instance:wordpressdev
	return provider.Namespace + ":" + provider.Name + ":" + resourceType + ":" + resourceName
}

// ProviderExtensions returns the extended information of provider based on
// the provider and type of the resource.
func ProviderExtensions(provider *inputs.Provider, providerMeta map[string]any, resourceType string) map[string]interface{} {
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

// AppendToIntent adds a Kubernetes resource to the Intent resources slice.
func AppendToIntent(resourceType intent.Type, resourceID string, i *intent.Intent, resource any) error {
	// this function is only used for Kubernetes resources
	if resourceType != intent.Kubernetes {
		return errors.New("AppendToIntent is only used for Kubernetes resources")
	}

	gvk := resource.(runtime.Object).GetObjectKind().GroupVersionKind().String()
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	r := intent.Resource{
		ID:         resourceID,
		Type:       resourceType,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: map[string]any{
			intent.ResourceExtensionGVK: gvk,
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
func PatchResource[T any](resources map[string][]*intent.Resource, gvk string, patchFunc func(*T) error) error {
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

// AddKubeConfigIf adds kubeConfig from workspace to extensions of Kubernetes type resource in intent.
// If there is already has kubeConfig in extensions, use the kubeConfig in extensions.
func AddKubeConfigIf(i *intent.Intent, ws *workspaceapi.Workspace) {
	config, err := workspace.GetKubernetesConfig(ws.Runtimes)
	if errors.Is(err, workspace.ErrEmptyRuntimeConfigs) || errors.Is(err, workspace.ErrEmptyKubernetesConfig) {
		return
	}
	kubeConfig := config.KubeConfig
	if kubeConfig == "" {
		return
	}
	for n, resource := range i.Resources {
		if resource.Type == intent.Kubernetes {
			if resource.Extensions == nil {
				i.Resources[n].Extensions = make(map[string]any)
			}
			if extensionsKubeConfig, ok := resource.Extensions[intent.ResourceExtensionKubeConfig]; !ok || extensionsKubeConfig == "" {
				i.Resources[n].Extensions[intent.ResourceExtensionKubeConfig] = kubeConfig
			}
		}
	}
}
