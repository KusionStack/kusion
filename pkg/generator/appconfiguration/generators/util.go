package generators

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
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
func ForeachOrdered[T any](
	m map[string]T,
	f func(key string, value T) error,
) error {
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

// IntPtr returns a pointer to the provided value.
func IntPtr[T any](i T) *T {
	return &i
}

// kubernetesResourceID returns the unique ID of a Kubernetes resource
// based on its type and metadata.
func kubernetesResourceID(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:code-city:code-citydev
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id += objectMeta.Namespace + ":"
	}
	id += objectMeta.Name
	return id
}

// appendToSpec adds a Kubernetes resource to a spec's resources
// slice.
func appendToSpec(resourceID string, resource any, spec *models.Spec) error {
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	r := models.Resource{
		ID:         resourceID,
		Type:       generator.Kubernetes,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: nil,
	}
	spec.Resources = append(spec.Resources, r)
	return nil
}

// uniqueWorkloadName returns a unique name for a workload based on
// its project and name.
func uniqueWorkloadName(projectName, compName string) string {
	return projectName + "-" + compName
}

// uniqueWorkloadLabels returns a map of labels that identify a
// app based on its project and name.
func uniqueWorkloadLabels(projectName, appName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of": projectName,
		"app.kubernetes.io/name":    appName,
	}
}

func toOrderedContainers(appContainers map[string]container.Container) ([]corev1.Container, error) {
	// Create a slice of containers based on the app's
	// containers.
	containers := []corev1.Container{}
	if err := ForeachOrdered(appContainers, func(containerName string, c container.Container) error {
		// Create a slice of env vars based on the container's
		// envvars.
		envs := []corev1.EnvVar{}
		for k, v := range c.Env {
			envs = append(envs, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}

		// Create a container object and append it to the containers
		// slice.
		containers = append(containers, corev1.Container{
			Name:       containerName,
			Image:      c.Image,
			Command:    c.Command,
			Args:       c.Args,
			WorkingDir: c.WorkingDir,
			Env:        envs,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return containers, nil
}
