package workload

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

// workloadServiceGenerator is a struct for generating service
// workload resources.
type workloadServiceGenerator struct {
	project *projectstack.Project
	stack   *projectstack.Stack
	appName string
	service *workload.Service
}

// NewWorkloadServiceGenerator returns a new workloadServiceGenerator
// instance.
func NewWorkloadServiceGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	service *workload.Service,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if service == nil {
		return nil, fmt.Errorf("service workload must not be nil")
	}

	return &workloadServiceGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		service: service,
	}, nil
}

// NewWorkloadServiceGeneratorFunc returns a new NewGeneratorFunc that
// returns a workloadServiceGenerator instance.
func NewWorkloadServiceGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	service *workload.Service,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewWorkloadServiceGenerator(project, stack, appName, service)
	}
}

// Generate generates a service workload resource to the given spec.
func (g *workloadServiceGenerator) Generate(spec *models.Spec) error {
	service := g.service
	if service == nil {
		return nil
	}

	// Create an empty resource slice if it doesn't exist yet.
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	// Create a slice of containers based on the app's
	// containers.
	containers, err := toOrderedContainers(service.Containers)
	if err != nil {
		return err
	}

	// Create a Deployment object based on the app's
	// configuration.
	resource := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: appconfiguration.MergeMaps(
				appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
				g.service.Labels,
			),
			Annotations: appconfiguration.MergeMaps(
				g.service.Annotations,
			),
			Name:      appconfiguration.UniqueAppName(g.project.Name, g.stack.Name, g.appName),
			Namespace: g.project.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: appconfiguration.GenericPtr(int32(service.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: appconfiguration.MergeMaps(
						appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
						g.service.Labels,
					),
					Annotations: appconfiguration.MergeMaps(
						g.service.Annotations,
					),
				},
				Spec: v1.PodSpec{
					Containers: containers,
				},
			},
		},
	}

	// Add the Deployment resource to the spec.
	return appconfiguration.AppendToSpec(
		appconfiguration.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta),
		resource,
		spec,
	)
}
