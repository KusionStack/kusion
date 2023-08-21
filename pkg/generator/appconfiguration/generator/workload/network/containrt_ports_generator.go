package network

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	ac "kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
	"kusionstack.io/kusion/pkg/projectstack"
)

const kindK8sService = "Service"

// containerPortsGenerator is used to generate k8s service.
type containerPortsGenerator struct {
	project        *projectstack.Project
	stack          *projectstack.Stack
	appName        string
	workloadLabels map[string]string
	containerPorts []network.ContainerPort
}

// NewContainerPortsGenerator returns a new containerPortsGenerator instance.
func NewContainerPortsGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workloadLabels map[string]string,
	cps []network.ContainerPort,
) (ac.Generator, error) {
	if project.Name == "" {
		return nil, fmt.Errorf("project name must not be empty")
	}
	if stack.Name == "" {
		return nil, fmt.Errorf("stack name must not be empty")
	}
	if appName == "" {
		return nil, fmt.Errorf("app name must not be empty")
	}
	if len(cps) == 0 {
		return nil, fmt.Errorf("container posts must not be empty")
	}

	return &containerPortsGenerator{
		project:        project,
		stack:          stack,
		appName:        appName,
		workloadLabels: workloadLabels,
		containerPorts: cps,
	}, nil
}

// NewContainerPortsGeneratorFunc returns a new NewGeneratorFunc that returns a
// containerPortsGenerator instance.
func NewContainerPortsGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workloadLabels map[string]string,
	cps []network.ContainerPort,
) ac.NewGeneratorFunc {
	return func() (ac.Generator, error) {
		return NewContainerPortsGenerator(project, stack, appName, workloadLabels, cps)
	}
}

func (g *containerPortsGenerator) Generate(spec *models.Spec) error {
	var svcPorts []v1.ServicePort
	for _, cp := range g.containerPorts {
		if cp.AccessMode == network.AccessModeExposed {
			svcPorts = append(svcPorts, v1.ServicePort{
				Port:       int32(cp.AccessPort),
				TargetPort: intstr.FromInt(cp.Port),
				Protocol:   v1.Protocol(cp.AccessProtocol),
			})
		}
	}
	// no port should get exposed
	if len(svcPorts) == 0 {
		return nil
	}

	resource := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       kindK8sService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ac.UniqueAppName(g.project.Name, g.stack.Name, g.appName),
			Namespace: g.project.Name,
		},
		Spec: v1.ServiceSpec{
			Ports: svcPorts,
			Selector: ac.MergeMaps(
				ac.UniqueAppLabels(g.project.Name, g.appName),
				g.workloadLabels,
			),
		},
	}
	resourceID := ac.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta)

	return ac.AppendToSpec(resourceID, resource, spec)
}
