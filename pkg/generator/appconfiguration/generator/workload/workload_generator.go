package workload

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
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

		switch g.workload.Header.Type {
		case workload.TypeService:
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g.project, g.stack, g.appName, g.workload.Service))
		case workload.TypeJob:
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

		// Create a container object and append it to the containers slice.
		containers = append(containers, corev1.Container{
			Name:       containerName,
			Image:      c.Image,
			Command:    c.Command,
			Args:       c.Args,
			WorkingDir: c.WorkingDir,
			Env:        envs,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return containers, nil
}
