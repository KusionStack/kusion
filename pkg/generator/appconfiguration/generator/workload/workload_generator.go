package workload

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
	networkmodel "kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
	"kusionstack.io/kusion/pkg/projectstack"
)

type workloadGenerator struct {
	project  *projectstack.Project
	stack    *projectstack.Stack
	appName  string
	workload *workload.Workload
}

func NewWorkloadGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	workload *workload.Workload,
	appName string,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &workloadGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
	}, nil
}

func NewWorkloadGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	workload *workload.Workload,
	appName string,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewWorkloadGenerator(project, stack, workload, appName)
	}
}

func (g *workloadGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	if g.workload != nil {
		var gfs []appconfiguration.NewGeneratorFunc

		switch g.workload.Type {
		case workload.WorkloadTypeService:
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g.project, g.stack, g.appName, g.workload.Service))
		case workload.WorkloadTypeJob:
			gfs = append(gfs, NewJobGeneratorFunc(g.project, g.stack, g.appName, g.workload.Job))
		}

		if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}

func toOrderedContainers(appContainers map[string]container.Container) ([]corev1.Container, error) {
	// Create a slice of containers based on the app's
	// containers.
	var containers []corev1.Container
	if err := appconfiguration.ForeachOrdered(appContainers, func(containerName string, c container.Container) error {
		// Create a slice of env vars based on the container's env vars.
		var envs []corev1.EnvVar
		for k, v := range c.Env {
			envs = append(envs, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}

		// Generate a slice of k8s containerPort.
		var cps []corev1.ContainerPort
		for _, port := range c.Ports {
			cps = append(cps, corev1.ContainerPort{
				ContainerPort: int32(port.Port),
				Protocol:      corev1.Protocol(port.AccessProtocol),
			})
		}

		// Create a container object and append it to the containers slice.
		containers = append(containers, corev1.Container{
			Name:       containerName,
			Image:      c.Image,
			Command:    c.Command,
			Args:       c.Args,
			WorkingDir: c.WorkingDir,
			Env:        envs,
			Ports:      cps,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return containers, nil
}

func completeContainerPorts(appContainers map[string]container.Container) {
	for _, c := range appContainers {
		for _, cp := range c.Ports {
			cp.Complete()
		}
	}
}

func validContainerPorts(appContainers map[string]container.Container) error {
	fmtErr := func(port int, portType, name1, name2 string) error {
		if name1 == name2 {
			return fmt.Errorf("more than one %s is %d in container %s",
				portType, port, name1)
		} else {
			return fmt.Errorf("container %s and %s have same %s %d",
				name1, name2, portType, port)
		}
	}

	portName := make(map[int]string)
	accessPortName := make(map[int]string)
	for name, c := range appContainers {
		for _, cp := range c.Ports {
			if portName[cp.Port] != "" {
				portName[cp.Port] = name
			} else {
				return fmtErr(cp.Port, networkmodel.Port, portName[cp.Port], name)
			}
			if accessPortName[cp.AccessPort] != "" {
				accessPortName[cp.AccessPort] = name
			} else {
				return fmtErr(cp.AccessPort, networkmodel.AccessPort, accessPortName[cp.AccessPort], name)
			}
		}
	}

	return nil
}

func toContainerPorts(appContainers map[string]container.Container) []*networkmodel.ContainerPort {
	var cps []*networkmodel.ContainerPort
	for _, c := range appContainers {
		for _, cp := range c.Ports {
			cps = append(cps, &cp)
		}
	}
	return cps
}
