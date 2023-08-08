package generators

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
)

// deploymentGenerator is a struct for generating Deployment
// resources.
type deploymentGenerator struct {
	projectName string
	compName    string
	comp        *component.Component
}

// NewDeploymentGenerator returns a new deploymentGenerator instance.
func NewDeploymentGenerator(
	projectName string,
	compName string,
	comp *component.Component,
) (Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(compName) == 0 {
		return nil, fmt.Errorf("component name must not be empty")
	}

	if comp == nil {
		return nil, fmt.Errorf("component must not be nil")
	}

	return &deploymentGenerator{
		projectName: projectName,
		compName:    compName,
		comp:        comp,
	}, nil
}

// NewDeploymentGeneratorFunc returns a new NewGeneratorFunc that
// returns a deploymentGenerator instance.
func NewDeploymentGeneratorFunc(
	projectName string,
	compName string,
	comp *component.Component,
) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewDeploymentGenerator(projectName, compName, comp)
	}
}

// Generate generates a Deployment resource to the given spec.
func (g *deploymentGenerator) Generate(spec *models.Spec) error {
	lrs := g.comp.LongRunningService
	if lrs == nil {
		return nil
	}

	// Create an empty resource slice if it doesn't exist yet.
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	// Create a slice of containers based on the component's
	// containers.
	containers, err := toOrderedContainers(lrs.Containers)
	if err != nil {
		return err
	}

	// Create a Deployment object based on the component's
	// configuration.
	resource := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    uniqueComponentLabels(g.projectName, g.compName),
			Name:      uniqueComponentName(g.projectName, g.compName),
			Namespace: g.projectName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(lrs.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: uniqueComponentLabels(g.projectName, g.compName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: uniqueComponentLabels(g.projectName, g.compName),
				},
				Spec: v1.PodSpec{
					Containers: containers,
				},
			},
		},
	}

	// Add the Deployment resource to the spec.
	return appendToSpec(
		kubernetesResourceID(resource.TypeMeta, resource.ObjectMeta),
		resource,
		spec,
	)
}
