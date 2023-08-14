package generators

import (
	"fmt"

	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
)

type workloadGenerator struct {
	projectName string
	appName     string
	workload    *workload.Workload
}

func NewWorkloadGenerator(projectName, appName string, workload *workload.Workload) (Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &workloadGenerator{
		projectName: projectName,
		appName:     appName,
		workload:    workload,
	}, nil
}

func NewWorkloadGeneratorFunc(projectName, appName string, workload *workload.Workload) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewWorkloadGenerator(projectName, appName, workload)
	}
}

func (g *workloadGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	if g.workload != nil {
		gfs := []NewGeneratorFunc{}

		switch g.workload.Type {
		case workload.WorkloadTypeService:
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g.projectName, g.appName, g.workload.Service))
		case workload.WorkloadTypeJob:
			gfs = append(gfs, NewJobGeneratorFunc(g.projectName, g.appName, g.workload.Job))
		}

		if err := CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}
