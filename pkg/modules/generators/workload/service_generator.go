package workload

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kube-api/apps/v1alpha1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload/network"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/workspace"
)

// workloadServiceGenerator is a struct for generating service workload resources.
type workloadServiceGenerator struct {
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	service       *workload.Service
	serviceConfig apiv1.GenericConfig
	namespace     string

	// for internal generator
	context modules.GeneratorContext
}

// NewWorkloadServiceGenerator returns a new workloadServiceGenerator instance.
func NewWorkloadServiceGenerator(ctx modules.GeneratorContext) (modules.Generator, error) {
	if len(ctx.Project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(ctx.Application.Name) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if ctx.Application.Workload.Service == nil {
		return nil, fmt.Errorf("service workload must not be nil")
	}

	return &workloadServiceGenerator{
		project:       ctx.Project,
		stack:         ctx.Stack,
		appName:       ctx.Application.Name,
		service:       ctx.Application.Workload.Service,
		serviceConfig: ctx.ModuleInputs[workload.ModuleService],
		namespace:     ctx.Namespace,
		context:       ctx,
	}, nil
}

// NewWorkloadServiceGeneratorFunc returns a new NewGeneratorFunc that returns a workloadServiceGenerator instance.
func NewWorkloadServiceGeneratorFunc(ctx modules.GeneratorContext) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewWorkloadServiceGenerator(ctx)
	}
}

// Generate generates a service workload resource to the given spec.
func (g *workloadServiceGenerator) Generate(spec *apiv1.Intent) error {
	service := g.service
	if service == nil {
		return nil
	}

	// Create an empty resource slice if it doesn't exist yet.
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	if err := completeServiceInput(g.service, g.serviceConfig); err != nil {
		return fmt.Errorf("complete service input by workspace config failed, %w", err)
	}

	uniqueAppName := modules.UniqueAppName(g.project.Name, g.stack.Name, g.appName)

	// Create a slice of containers based on the app's
	// containers along with related volumes and configMaps.
	containers, volumes, configMaps, err := toOrderedContainers(service.Containers, uniqueAppName)
	if err != nil {
		return err
	}

	// Create ConfigMap objects based on the app's configuration.
	for _, cm := range configMaps {
		cmObj := cm
		cmObj.Namespace = g.namespace
		if err = modules.AppendToIntent(
			apiv1.Kubernetes,
			modules.KubernetesResourceID(cmObj.TypeMeta, cmObj.ObjectMeta),
			spec,
			&cmObj,
		); err != nil {
			return err
		}
	}

	labels := modules.MergeMaps(modules.UniqueAppLabels(g.project.Name, g.appName), g.service.Labels)
	annotations := modules.MergeMaps(g.service.Annotations)
	selector := modules.UniqueAppLabels(g.project.Name, g.appName)

	// Create a K8s workload object based on the app's configuration.
	// common parts
	objectMeta := metav1.ObjectMeta{
		Labels:      labels,
		Annotations: annotations,
		Name:        uniqueAppName,
		Namespace:   g.namespace,
	}
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: containers,
			Volumes:    volumes,
		},
	}

	var resource any
	typeMeta := metav1.TypeMeta{}

	switch service.Type {
	case workload.TypeDeployment:
		typeMeta = metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       workload.TypeDeployment,
		}
		spec := appsv1.DeploymentSpec{
			Replicas: modules.GenericPtr(int32(service.Replicas)),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: podTemplateSpec,
		}
		resource = &appsv1.Deployment{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec:       spec,
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
				Replicas: modules.GenericPtr(int32(service.Replicas)),
				Selector: &metav1.LabelSelector{MatchLabels: selector},
				Template: podTemplateSpec,
			},
		}
	}

	// Add the Deployment resource to the spec.
	if err = modules.AppendToIntent(apiv1.Kubernetes, modules.KubernetesResourceID(typeMeta, objectMeta), spec, resource); err != nil {
		return err
	}

	// generate K8s Service from ports config.
	if len(g.service.Ports) != 0 {
		portsGeneratorFunc := network.NewPortsGeneratorFunc(g.context, selector, labels, annotations)
		if err = modules.CallGenerators(spec, portsGeneratorFunc); err != nil {
			return err
		}
	}

	return nil
}

func completeServiceInput(service *workload.Service, config apiv1.GenericConfig) error {
	if err := completeBaseWorkload(&service.Base, config); err != nil {
		return err
	}
	serviceType, err := workspace.GetStringFromGenericConfig(config, workload.FieldType)
	if err != nil {
		return err
	}
	// if not set in workspace, use Deployment as default type
	if serviceType == "" {
		serviceType = workload.TypeDeployment
	}
	if serviceType != workload.TypeDeployment && serviceType != workload.TypeCollaset {
		return fmt.Errorf("unsupported service type %s", serviceType)
	}
	if service.Type == "" {
		service.Type = serviceType
	}
	return nil
}
