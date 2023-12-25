package network

import (
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/network"
	"kusionstack.io/kusion/pkg/workspace"
)

const (
	k8sKindService = "Service"
	suffixPublic   = "public"
	suffixPrivate  = "private"
)

var (
	ErrEmptyAppName              = errors.New("app name must not be empty")
	ErrEmptyProjectName          = errors.New("project name must not be empty")
	ErrEmptyStackName            = errors.New("stack name must not be empty")
	ErrEmptySelectors            = errors.New("selectors must not be empty")
	ErrEmptyPorts                = errors.New("ports must not be empty")
	ErrEmptyType                 = errors.New("type must not be empty when public")
	ErrUnsupportedType           = errors.New("type only support alicloud and aws for now")
	ErrInvalidPort               = errors.New("port must be between 1 and 65535")
	ErrInvalidTargetPort         = errors.New("targetPort must be between 1 and 65535 if exist")
	ErrInvalidProtocol           = errors.New("protocol must be TCP or UDP")
	ErrDuplicatePortProtocol     = errors.New("port-protocol pair must not be duplicate")
	ErrUnsupportedPortConfigItem = errors.New("unsupported item for port workspace config")
	ErrEmptyPortConfig           = errors.New("empty port config")
)

// portsGenerator is used to generate k8s service.
type portsGenerator struct {
	appName     string
	projectName string
	stackName   string
	selector    map[string]string
	labels      map[string]string
	annotations map[string]string
	ports       []network.Port
	portConfig  apiv1.GenericConfig
	namespace   string
}

// NewPortsGenerator returns a new portsGenerator instance, and do the validation and completion job.
func NewPortsGenerator(
	ctx modules.GeneratorContext,
	selectors, labels, annotations map[string]string,
) (modules.Generator, error) {
	generator := &portsGenerator{
		appName:     ctx.Application.Name,
		projectName: ctx.Project.Name,
		stackName:   ctx.Stack.Name,
		selector:    selectors,
		labels:      labels,
		annotations: annotations,
		ports:       ctx.Application.Workload.Service.Ports,
		portConfig:  ctx.ModuleInputs[network.ModulePort],
		namespace:   ctx.Namespace,
	}

	if err := generator.validate(); err != nil {
		return nil, err
	}
	if err := generator.complete(); err != nil {
		return nil, err
	}

	return generator, nil
}

// NewPortsGeneratorFunc returns a new NewGeneratorFunc that returns a portsGenerator instance.
func NewPortsGeneratorFunc(
	ctx modules.GeneratorContext,
	selectors, labels, annotations map[string]string,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewPortsGenerator(ctx, selectors, labels, annotations)
	}
}

// Generate renders k8s ClusterIP or LoadBalancer service from the portsGenerator.
func (g *portsGenerator) Generate(spec *apiv1.Intent) error {
	privatePorts, publicPorts := splitPorts(g.ports)
	if len(privatePorts) != 0 {
		svc := g.generateK8sSvc(false, privatePorts)
		if err := appendToSpec(spec, svc); err != nil {
			return err
		}
	}
	if len(publicPorts) != 0 {
		svc := g.generateK8sSvc(true, publicPorts)
		if err := appendToSpec(spec, svc); err != nil {
			return err
		}
	}
	return nil
}

func (g *portsGenerator) validate() error {
	if g.appName == "" {
		return ErrEmptyAppName
	}
	if g.projectName == "" {
		return ErrEmptyProjectName
	}
	if g.stackName == "" {
		return ErrEmptyStackName
	}
	if len(g.selector) == 0 {
		return ErrEmptySelectors
	}
	if len(g.ports) == 0 {
		return ErrEmptyPorts
	}
	if err := validatePorts(g.ports); err != nil {
		return err
	}
	if err := validatePortConfig(g.portConfig); err != nil {
		return err
	}
	return nil
}

func (g *portsGenerator) complete() error {
	for i := range g.ports {
		if err := completePort(&g.ports[i], g.portConfig); err != nil {
			return err
		}
	}
	return nil
}

func (g *portsGenerator) generateK8sSvc(public bool, ports []network.Port) *v1.Service {
	appUname := modules.UniqueAppName(g.projectName, g.stackName, g.appName)
	var name string
	if public {
		name = fmt.Sprintf("%s-%s", appUname, suffixPublic)
	} else {
		name = fmt.Sprintf("%s-%s", appUname, suffixPrivate)
	}
	svcType := v1.ServiceTypeClusterIP
	if public {
		svcType = v1.ServiceTypeLoadBalancer
	}

	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       k8sKindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   g.namespace,
			Labels:      g.labels,
			Annotations: g.annotations,
		},
		Spec: v1.ServiceSpec{
			Ports:    toSvcPorts(name, ports),
			Selector: g.selector,
			Type:     svcType,
		},
	}

	if public {
		if len(svc.Labels) == 0 {
			svc.Labels = make(map[string]string)
		}
		if len(svc.Annotations) == 0 {
			svc.Annotations = make(map[string]string)
		}

		labels := ports[0].Labels
		for k, v := range labels {
			svc.Labels[k] = v
		}
		annotations := ports[0].Annotations
		for k, v := range annotations {
			svc.Annotations[k] = v
		}
	}

	return svc
}

func validatePorts(ports []network.Port) error {
	portProtocolRecord := make(map[string]struct{})
	// portType is the correct type for public port, it gets assigned a value when calling validatePort.
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

func validatePort(port *network.Port) error {
	if port.Port < 1 || port.Port > 65535 {
		return ErrInvalidPort
	}
	if port.TargetPort < 0 || port.Port > 65535 {
		return ErrInvalidTargetPort
	}
	if port.Protocol != network.ProtocolTCP && port.Protocol != network.ProtocolUDP {
		return ErrInvalidProtocol
	}
	return nil
}

func validatePortConfig(portConfig apiv1.GenericConfig) error {
	if portConfig == nil {
		return nil
	}
	for k := range portConfig {
		if k != network.FieldType && k != network.FieldLabels && k != network.FieldAnnotations {
			return fmt.Errorf("%w, %s", ErrUnsupportedPortConfigItem, k)
		}
	}
	return nil
}

func completePort(port *network.Port, portConfig apiv1.GenericConfig) error {
	if port.TargetPort == 0 {
		port.TargetPort = port.Port
	}
	if port.Public {
		// get port type from workspace
		if portConfig == nil {
			return ErrEmptyPortConfig
		}
		portType, err := workspace.GetStringFromGenericConfig(portConfig, network.FieldType)
		if err != nil {
			return err
		}
		if portType == "" {
			return ErrEmptyType
		}
		if portType != network.CSPAliCloud && portType != network.CSPAWS {
			return ErrUnsupportedType
		}
		port.Type = portType

		// get labels from workspace
		labels, err := workspace.GetStringMapFromGenericConfig(portConfig, network.FieldLabels)
		if err != nil {
			return err
		}
		port.Labels = labels

		// get annotations from workspace
		annotations, err := workspace.GetStringMapFromGenericConfig(portConfig, network.FieldAnnotations)
		if err != nil {
			return err
		}
		port.Annotations = annotations
	}
	return nil
}

func splitPorts(ports []network.Port) ([]network.Port, []network.Port) {
	var privatePorts, publicPorts []network.Port
	for _, port := range ports {
		if port.Public {
			publicPorts = append(publicPorts, port)
		} else {
			privatePorts = append(privatePorts, port)
		}
	}
	return privatePorts, publicPorts
}

func toSvcPorts(name string, ports []network.Port) []v1.ServicePort {
	svcPorts := make([]v1.ServicePort, len(ports))
	for i, port := range ports {
		svcPorts[i] = v1.ServicePort{
			Name:       fmt.Sprintf("%s-%d-%s", name, port.Port, strings.ToLower(port.Protocol)),
			Port:       int32(port.Port),
			TargetPort: intstr.FromInt(port.TargetPort),
			Protocol:   v1.Protocol(port.Protocol),
		}
	}
	return svcPorts
}

func appendToSpec(spec *apiv1.Intent, svc *v1.Service) error {
	id := modules.KubernetesResourceID(svc.TypeMeta, svc.ObjectMeta)
	return modules.AppendToIntent(apiv1.Kubernetes, id, spec, svc)
}
