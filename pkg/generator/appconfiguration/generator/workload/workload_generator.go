package workload

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
)

type workloadGenerator struct {
	projectName string
	appName     string
	workload    *workload.Workload
}

func NewWorkloadGenerator(projectName, appName string, workload *workload.Workload) (appconfiguration.Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &workloadGenerator{
		projectName: projectName,
		appName:     appName,
		workload:    workload,
	}, nil
}

func NewWorkloadGeneratorFunc(projectName, appName string, workload *workload.Workload) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewWorkloadGenerator(projectName, appName, workload)
	}
}

func (g *workloadGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	if g.workload != nil {
		gfs := []appconfiguration.NewGeneratorFunc{}

		switch g.workload.Type {
		case workload.WorkloadTypeService:
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g.projectName, g.appName, g.workload.Service))
		case workload.WorkloadTypeJob:
			gfs = append(gfs, NewJobGeneratorFunc(g.projectName, g.appName, g.workload.Job))
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
	containers := []corev1.Container{}
	if err := appconfiguration.ForeachOrdered(appContainers, func(containerName string, c container.Container) error {
		// Create a slice of env vars based on the container's
		// envvars.
		envs := []corev1.EnvVar{}
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
