package builders

import (
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators"
	"kusionstack.io/kusion/pkg/modules/inputs"
)

type AppsConfigBuilder struct {
	Apps map[string]inputs.AppConfiguration
}

func (acg *AppsConfigBuilder) Build(
	_ *Options,
	project *project.Project,
	stack *stack.Stack,
) (*intent.Intent, error) {
	spec := &intent.Intent{
		Resources: []intent.Resource{},
	}

	var gfs []modules.NewGeneratorFunc
	err := modules.ForeachOrdered(acg.Apps, func(appName string, app inputs.AppConfiguration) error {
		gfs = append(gfs, generators.NewAppConfigurationGeneratorFunc(project, stack, appName, &app))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = modules.CallGenerators(spec, gfs...); err != nil {
		return nil, err
	}

	return spec, nil
}
