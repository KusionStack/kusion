// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builders

import (
	"fmt"

	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/generators"
	"kusionstack.io/kusion/pkg/generators/appconfiguration"
)

type AppsConfigBuilder struct {
	Apps      map[string]v1.AppConfiguration
	Workspace *v1.Workspace
}

func (acg *AppsConfigBuilder) Build(kclPackage *api.KclPackage, project *v1.Project, stack *v1.Stack) (*v1.Spec, error) {
	i := &v1.Spec{
		Resources: []v1.Resource{},
	}

	var gfs []generators.NewSpecGeneratorFunc
	err := generators.ForeachOrdered(acg.Apps, func(appName string, app v1.AppConfiguration) error {
		if kclPackage == nil {
			return fmt.Errorf("kcl package is nil when generating app configuration for %s", appName)
		}
		dependencies := kclPackage.GetDependenciesInModFile()
		gfs = append(gfs, appconfiguration.NewAppConfigurationGeneratorFunc(project, stack, appName, &app, acg.Workspace, dependencies))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = generators.CallGenerators(i, gfs...); err != nil {
		return nil, err
	}

	return i, nil
}
