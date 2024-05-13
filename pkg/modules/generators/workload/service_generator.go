package workload

import (
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kube-api/apps/v1alpha1"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/workspace"
)

var (
	ErrEmptySelectors        = errors.New("selectors must not be empty")
	ErrInvalidPort           = errors.New("port must be between 1 and 65535")
	ErrInvalidTargetPort     = errors.New("targetPort must be between 1 and 65535 if exist")
	ErrInvalidProtocol       = errors.New("protocol must be TCP or UDP")
	ErrDuplicatePortProtocol = errors.New("port-protocol pair must not be duplicate")
)

// ServiceGenerator is a struct for generating Service Workload resources.
type ServiceGenerator struct {
	Project   string
	Stack     string
	App       string
	Namespace string
	Service   *apiv1.Service
	Config    apiv1.GenericConfig
}

// NewWorkloadServiceGenerator returns a new ServiceGenerator instance.
func NewWorkloadServiceGenerator(request *Generator) (modules.Generator, error) {
	if len(request.Project) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(request.App) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if request.Workload.Service == nil {
		return nil, fmt.Errorf("service Workload must not be nil")
	}

	return &ServiceGenerator{
		Project:   request.Project,
		Stack:     request.Stack,
		App:       request.App,
		Service:   request.Workload.Service,
		Config:    request.PlatformConfigs[apiv1.ModuleService],
		Namespace: request.Namespace,
	}, nil
}

// NewWorkloadServiceGeneratorFunc returns a new NewGeneratorFunc that returns a ServiceGenerator instance.
func NewWorkloadServiceGeneratorFunc(workloadGenerator *Generator) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewWorkloadServiceGenerator(workloadGenerator)
	}
}

// Generate generates a Service Workload resource to the given spec.
func (g *ServiceGenerator) Generate(spec *apiv1.Spec) error {
	service := g.Service
	if service == nil {
		return nil
	}

	// Create an empty resource slice if it doesn't exist yet.
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	if err := completeServiceInput(g.Service, g.Config); err != nil {
		return fmt.Errorf("complete Service input by workspace config failed, %w", err)
	}

	uniqueAppName := modules.UniqueAppName(g.Project, g.Stack, g.App)

	// Create a slice of containers based on the App's
	// containers along with related volumes and configMaps.
	containers, volumes, configMaps, err := toOrderedContainers(service.Containers, uniqueAppName)
	if err != nil {
		return err
	}

	// Create ConfigMap objects based on the App's configuration.
	for _, cm := range configMaps {
		cmObj := cm
		cmObj.Namespace = g.Namespace
		if err = modules.AppendToSpec(
			apiv1.Kubernetes,
			modules.KubernetesResourceID(cmObj.TypeMeta, cmObj.ObjectMeta),
			spec,
			&cmObj,
		); err != nil {
			return err
		}
	}

	labels := modules.MergeMaps(modules.UniqueAppLabels(g.Project, g.App), g.Service.Labels)
	annotations := modules.MergeMaps(g.Service.Annotations)
	selectors := modules.UniqueAppLabels(g.Project, g.App)

	// Create a K8s Workload object based on the App's configuration.
	// common parts
	objectMeta := metav1.ObjectMeta{
		Labels:      labels,
		Annotations: annotations,
		Name:        uniqueAppName,
		Namespace:   g.Namespace,
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
	case apiv1.Deployment:
		typeMeta = metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       string(apiv1.Deployment),
		}
		spec := appsv1.DeploymentSpec{
			Replicas: service.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectors},
			Template: podTemplateSpec,
		}
		resource = &appsv1.Deployment{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec:       spec,
		}
	case apiv1.Collaset:
		typeMeta = metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.String(),
			Kind:       string(apiv1.Collaset),
		}
		resource = &v1alpha1.CollaSet{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec: v1alpha1.CollaSetSpec{
				Replicas: service.Replicas,
				Selector: &metav1.LabelSelector{MatchLabels: selectors},
				Template: podTemplateSpec,
			},
		}
	}

	// Add the Deployment resource to the spec.
	if err = modules.AppendToSpec(apiv1.Kubernetes, modules.KubernetesResourceID(typeMeta, objectMeta), spec, resource); err != nil {
		return err
	}

	// validate and complete service ports
	if len(g.Service.Ports) != 0 {
		if err = validate(selectors, service.Ports); err != nil {
			return err
		}
		if err = complete(service.Ports); err != nil {
			return err
		}
	}
	return nil
}

func validatePorts(ports []apiv1.Port) error {
	portProtocolRecord := make(map[string]struct{})
	for _, port := range ports {
		if err := validatePort(&port); err != nil {
			return fmt.Errorf("invalid port config %+v, %w", port, err)
		}

		// duplicate "port-protocol" pairs are not allowed.
		portProtocol := fmt.Sprintf("%d-%s", port.Port, port.Protocol)
		if _, ok := portProtocolRecord[portProtocol]; ok {
			return fmt.Errorf("invalid port config %+v, %v", port, ErrDuplicatePortProtocol)
		}
		portProtocolRecord[portProtocol] = struct{}{}
	}
	return nil
}

func validatePort(port *apiv1.Port) error {
	if port.Port < 1 || port.Port > 65535 {
		return ErrInvalidPort
	}
	if port.TargetPort < 0 || port.Port > 65535 {
		return ErrInvalidTargetPort
	}
	if port.Protocol != apiv1.TCP && port.Protocol != apiv1.UDP {
		return ErrInvalidProtocol
	}
	return nil
}

func validate(selectors map[string]string, ports []apiv1.Port) error {
	if len(selectors) == 0 {
		return ErrEmptySelectors
	}
	if err := validatePorts(ports); err != nil {
		return err
	}
	return nil
}

func complete(ports []apiv1.Port) error {
	for i := range ports {
		if ports[i].TargetPort == 0 {
			ports[i].TargetPort = ports[i].Port
		}
	}
	return nil
}

func completeServiceInput(service *apiv1.Service, config apiv1.GenericConfig) error {
	if err := completeBaseWorkload(&service.Base, config); err != nil {
		return err
	}
	serviceTypeStr, err := workspace.GetStringFromGenericConfig(config, apiv1.ModuleServiceType)
	platformServiceType := apiv1.ServiceType(serviceTypeStr)
	if err != nil {
		return err
	}
	// if not set in workspace, use Deployment as default type
	if platformServiceType == "" {
		platformServiceType = apiv1.Deployment
	}
	if platformServiceType != apiv1.Deployment && platformServiceType != apiv1.Collaset {
		return fmt.Errorf("unsupported Service type %s", platformServiceType)
	}
	if service.Type == "" {
		service.Type = platformServiceType
	}
	return nil
}
