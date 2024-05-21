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

package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"kcl-lang.io/kpm/pkg/api"
	"kcl-lang.io/kpm/pkg/env"
	pkg "kcl-lang.io/kpm/pkg/package"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/api/builders"
	"kusionstack.io/kusion/pkg/engine/api/generate/run"
	"kusionstack.io/kusion/pkg/modules/generators"
	"kusionstack.io/kusion/pkg/util/io"
	"kusionstack.io/kusion/pkg/util/kfile"
)

// Generator is an interface for things that can generate versioned Spec from
// configuration code under current working directory with given input parameters.
type Generator interface {
	// Generate creates versioned Intent given working directory and a set of parameters
	Generate(workDir string, params map[string]string) (*v1.Spec, error)
}

// DefaultGenerator is the default Generator implementation.
type DefaultGenerator struct {
	Project   *v1.Project
	Stack     *v1.Stack
	Workspace *v1.Workspace
	Runner    run.CodeRunner
}

// Generate versioned Spec with target code runner.
func (g *DefaultGenerator) Generate(workDir string, params map[string]string) (*v1.Spec, error) {
	// Call code runner to generate raw data
	if params == nil {
		params = make(map[string]string, 1)
	}
	rawAppConfiguration, err := g.Runner.Run(workDir, params)
	if err != nil {
		return nil, err
	}

	// Copy dependent modules before call builder
	err = CopyDependentModules(workDir)
	if err != nil {
		return nil, err
	}

	// Note: we use the type of MapSlice in yaml.v2 to maintain the order of container
	// environment variables, thus we unmarshal appConfigs with yaml.v2 rather than yaml.v3.
	apps := map[string]v1.AppConfiguration{}
	err = yaml.Unmarshal(rawAppConfiguration, apps)
	if err != nil {
		return nil, err
	}

	kclPkg, err := api.GetKclPackage(g.Stack.Path)
	if err != nil {
		return nil, err
	}

	builder := &builders.AppsConfigBuilder{
		Workspace: g.Workspace,
		Apps:      apps,
	}
	return builder.Build(kclPkg, g.Project, g.Stack)
}

// CopyDependentModules copies dependent Kusion modules' generators to destination.
func CopyDependentModules(workDir string) error {
	modFile := &pkg.ModFile{}
	err := modFile.LoadModFile(filepath.Join(workDir, pkg.MOD_FILE))
	if err != nil {
		return fmt.Errorf("load kcl.mod failed: %v", err)
	}

	absPkgPath, _ := env.GetAbsPkgPath()
	kusionHomePath, _ := kfile.KusionDataFolder()

	var allErrs []error
	for _, dep := range modFile.Deps {
		if dep.Source.Oci != nil {
			// ignore workload modules
			if generators.IgnoreModules[dep.Name] {
				continue
			}

			info := dep.Source.Oci
			pkgDir := filepath.Join(absPkgPath, dep.FullName)
			source := filepath.Join(pkgDir, "kusion-module-"+dep.FullName)
			moduleDir := filepath.Join(kusionHomePath, "modules", info.Repo, info.Tag, runtime.GOOS, runtime.GOARCH)
			dest := filepath.Join(moduleDir, fmt.Sprintf("kusion-module-%s", dep.FullName))
			if runtime.GOOS == "windows" {
				source = fmt.Sprintf("%s.exe", source)
				dest = fmt.Sprintf("%s.exe", dest)
			}
			// copy the module binary to the $KUSION_HOME modules directory
			err = io.CopyFile(source, dest)
			if err == nil {
				// mark the dest file as executable
				err = os.Chmod(dest, 0o755)
			}
			allErrs = append(allErrs, err)
		}
	}

	if allErrs != nil {
		return utilerrors.NewAggregate(allErrs)
	}

	return nil
}
