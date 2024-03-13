package builders

import (
	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators"
)

type AppsConfigBuilder struct {
	Apps      map[string]v1.AppConfiguration
	Workspace *v1.Workspace
}

func (acg *AppsConfigBuilder) Build(
	_ *Options,
	project *v1.Project,
	stack *v1.Stack,
) (*v1.Intent, error) {
	i := &v1.Intent{
		Resources: []v1.Resource{},
	}

	var gfs []modules.NewGeneratorFunc
	err := modules.ForeachOrdered(acg.Apps, func(appName string, app v1.AppConfiguration) error {
		gfs = append(gfs, generators.NewAppConfigurationGeneratorFunc(project.Name, stack.Name, appName, &app, acg.Workspace))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = modules.CallGenerators(i, gfs...); err != nil {
		return nil, err
	}

	return i, nil
}
