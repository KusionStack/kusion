package builders

import (
	"fmt"

	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	internalv1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators"
)

type AppsConfigBuilder struct {
	Apps      map[string]internalv1.AppConfiguration
	Workspace *v1.Workspace
}

func (acg *AppsConfigBuilder) Build(kclPackage *api.KclPackage, project *v1.Project, stack *v1.Stack) (*v1.Spec, error) {
	i := &v1.Spec{
		Resources: []v1.Resource{},
	}

	var gfs []modules.NewGeneratorFunc
	err := modules.ForeachOrdered(acg.Apps, func(appName string, app internalv1.AppConfiguration) error {
		if kclPackage == nil {
			return fmt.Errorf("kcl package is nil when generating app configuration for %s", appName)
		}
		dependencies := kclPackage.GetDependenciesInModFile()
		gfs = append(gfs, generators.NewAppConfigurationGeneratorFunc(project.Name, stack.Name, appName, &app, acg.Workspace, dependencies))
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
