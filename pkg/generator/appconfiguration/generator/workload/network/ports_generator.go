package network

import (
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	ac "kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
)

const (
	k8sKindService = "Service"
	suffixPublic   = "public"
	suffixPrivate  = "private"

	// aliyun SLB annotations, ref: https://help.aliyun.com/zh/ack/ack-managed-and-ack-dedicated/user-guide/add-annotations-to-the-yaml-file-of-a-service-to-configure-clb-instances
	aliyunLBSpec     = "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec"
	aliyunSLBS1Small = "slb.s1.small"

	// the label used for KafeD service controller
	kusionControl = "kusionstack.io/control"
)

var (
	ErrEmptyAppName          = errors.New("app name must not be empty")
	ErrEmptyProjectName      = errors.New("project name must not be empty")
	ErrEmptyStackName        = errors.New("stack name must not be empty")
	ErrEmptySelectors        = errors.New("selectors must not be empty")
	ErrEmptyPorts            = errors.New("ports must not be empty")
	ErrEmptyType             = errors.New("type must not be empty when public")
	ErrUnsupportedType       = errors.New("type only support aliyun and aws for now")
	ErrInconsistentType      = errors.New("public ports must use same type")
	ErrInvalidPort           = errors.New("port must be between 1 and 65535")
	ErrInvalidTargetPort     = errors.New("targetPort must be between 1 and 65535 if exist")
	ErrInvalidProtocol       = errors.New("protocol must be TCP or UDP")
	ErrDuplicatePortProtocol = errors.New("port-protocol pair must not be duplicate")
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
}

// NewPortsGenerator returns a new portsGenerator instance, and do the validation and completion job.
func NewPortsGenerator(
	appName, projectName, stackName string,
	selectors, labels, annotations map[string]string,
	ports []network.Port,
) (ac.Generator, error) {
	generator := &portsGenerator{
		appName:     appName,
		projectName: projectName,
		stackName:   stackName,
		selector:    selectors,
		labels:      labels,
		annotations: annotations,
		ports:       ports,
	}

	if err := generator.validate(); err != nil {
		return nil, err
	}

	generator.complete()
	return generator, nil
}

// NewPortsGeneratorFunc returns a new NewGeneratorFunc that returns a portsGenerator instance.
func NewPortsGeneratorFunc(
	appName, projectName, stackName string,
	selectors, labels, annotations map[string]string,
	ports []network.Port,
) ac.NewGeneratorFunc {
	return func() (ac.Generator, error) {
		return NewPortsGenerator(appName, projectName, stackName, selectors, labels, annotations, ports)
	}
}

// Generate renders k8s ClusterIP or LoadBalancer service from the portsGenerator.
func (g *portsGenerator) Generate(spec *models.Spec) error {
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
	return nil
}

func (g *portsGenerator) complete() {
	for i := range g.ports {
		completePort(&g.ports[i])
	}
}

func (g *portsGenerator) generateK8sSvc(public bool, ports []network.Port) *v1.Service {
	appUname := ac.UniqueAppName(g.projectName, g.stackName, g.appName)
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
			Namespace:   g.projectName,
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

		portType := ports[0].Type
		if portType == network.CSPAliyun {
			// for aliyun, set SLB spec by default.
			svc.Annotations[aliyunLBSpec] = aliyunSLBS1Small
			// kafeD service controller only support aliyun SLB, automatically add the label.
			svc.Labels[kusionControl] = "true"
		}
	}

	return svc
}

func validatePorts(ports []network.Port) error {
	portProtocolRecord := make(map[string]struct{})
	// portType is the correct type for public port, it gets assigned a value when calling validatePort.
	var portType string
	for _, port := range ports {
		if err := validatePort(&port, &portType); err != nil {
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

func validatePort(port *network.Port, portType *string) error {
	if port.Public {
		if port.Type == "" {
			return ErrEmptyType
		}
		if port.Type != network.CSPAliyun && port.Type != network.CSPAWS {
			return ErrUnsupportedType
		}
		if *portType == "" {
			*portType = port.Type
		} else if port.Type != *portType {
			return ErrInconsistentType
		}
	}
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

func completePort(port *network.Port) {
	if port.TargetPort == 0 {
		port.TargetPort = port.Port
	}
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

func appendToSpec(spec *models.Spec, svc *v1.Service) error {
	id := ac.KubernetesResourceID(svc.TypeMeta, svc.ObjectMeta)
	return ac.AppendToSpec(models.Kubernetes, id, spec, svc)
}
