package generators

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
)

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

// callGeneratorFuncs calls each NewGeneratorFunc in the given slice
// and returns a slice of Generator instances.
func callGeneratorFuncs(newGenerators ...NewGeneratorFunc) ([]Generator, error) {
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

// callGenerators calls the Generate method of each Generator instance
// returned by the given NewGeneratorFuncs.
func callGenerators(spec *models.Spec, newGenerators ...NewGeneratorFunc) error {
	gs, err := callGeneratorFuncs(newGenerators...)
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

// uniqueComponentName returns a unique name for a component based on
// its project and name.
func uniqueComponentName(projectName, compName string) string {
	return projectName + "-" + compName
}

// uniqueComponentLabels returns a map of labels that identify a
// component based on its project and name.
func uniqueComponentLabels(projectName, compName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      projectName,
		"app.kubernetes.io/component": compName,
	}
}

// int32Ptr returns a pointer to an int32 value.
func int32Ptr(i int32) *int32 {
	return &i
}
