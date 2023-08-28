package workload

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kube-api/apps/v1alpha1"

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

	uniqueAppName := appconfiguration.UniqueAppName(g.project.Name, g.stack.Name, g.appName)

	// Create a slice of containers based on the app's
	// containers along with related volumes and configMaps.
	containers, volumes, configMaps, err := toOrderedContainers(service.Containers, uniqueAppName)
	if err != nil {
		return err
	}

	// Create ConfigMap objects based on the app's configuration.
	for _, cm := range configMaps {
		cmObj := cm
		cmObj.Namespace = g.project.Name
		if err = appconfiguration.AppendToSpec(
			models.Kubernetes,
			appconfiguration.KubernetesResourceID(cmObj.TypeMeta, cmObj.ObjectMeta),
			spec,
			&cmObj,
		); err != nil {
			return err
		}
	}

	// Create a K8s workload object based on the app's configuration.
	// common parts
	objectMeta := metav1.ObjectMeta{
		Labels: appconfiguration.MergeMaps(
			appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
			g.service.Labels,
		),
		Annotations: appconfiguration.MergeMaps(
			g.service.Annotations,
		),
		Name:      uniqueAppName,
		Namespace: g.project.Name,
	}
	podTemplateSpec := v1.PodTemplateSpec{
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
			Volumes:    volumes,
		},
	}
	selector := &metav1.LabelSelector{
		MatchLabels: appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
	}

	var resource any
	typeMeta := metav1.TypeMeta{}

	switch service.Type {
	case workload.TypeDeploy:
		typeMeta = metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       workload.TypeDeploy,
		}
		resource = &appsv1.Deployment{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec: appsv1.DeploymentSpec{
				Replicas: appconfiguration.GenericPtr(int32(service.Replicas)),
				Selector: selector,
				Template: podTemplateSpec,
			},
		}
	case workload.TypeCollaset:
		typeMeta = metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.String(),
			Kind:       workload.TypeCollaset,
		}
		resource = &v1alpha1.CollaSet{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec: v1alpha1.CollaSetSpec{
				Replicas: appconfiguration.GenericPtr(int32(service.Replicas)),
				Selector: selector,
				Template: podTemplateSpec,
			},
		}
	}

	// Add the Deployment resource to the spec.
	return appconfiguration.AppendToSpec(models.Kubernetes, appconfiguration.KubernetesResourceID(typeMeta, objectMeta), spec, resource)
}
