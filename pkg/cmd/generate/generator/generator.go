package generator

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"kcl-lang.io/kpm/pkg/env"
	pkg "kcl-lang.io/kpm/pkg/package"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/cmd/generate/run"
	"kusionstack.io/kusion/pkg/util/io"
	"kusionstack.io/kusion/pkg/util/kfile"
)

const (
	IncludeSchemaTypePath = "include_schema_type_path"
)

// Generator is an interface for things that can generate versioned Intent from
// configuration code under current working directory and given input parameters.
type Generator interface {
	// Generate creates versioned Intent given working directory and set of parameters
	Generate(workDir string, params map[string]string) (*v1.Intent, error)
}

// DefaultGenerator is the default Generator implementation.
type DefaultGenerator struct {
	Project   *v1.Project
	Stack     *v1.Stack
	Workspace *v1.Workspace

	Runner run.CodeRunner
}

// Generate versioned Spec with target code runner.
func (g *DefaultGenerator) Generate(workDir string, params map[string]string) (*v1.Intent, error) {
	// Call code runner to generate raw data
	if params == nil {
		params = make(map[string]string, 1)
	}
	params[IncludeSchemaTypePath] = "true"
	rawAppConfiguration, err := g.Runner.Run(workDir, params)
	if err != nil {
		return nil, err
	}

	// Copy dependent modules before call builder
	err = copyDependentModules(workDir)
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

	builder := &builders.AppsConfigBuilder{
		Workspace: g.Workspace,
		Apps:      apps,
	}
	return builder.Build(nil, g.Project, g.Stack)
}

// copyDependentModules copies dependent Kusion modules' generators to destination.
func copyDependentModules(workDir string) error {
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
			info := dep.Source.Oci
			pkgDir := filepath.Join(absPkgPath, dep.FullName)
			platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
			source := filepath.Join(pkgDir, "_dist", platform, "generator")
			moduleDir := filepath.Join(kusionHomePath, "modules", info.Repo, info.Tag, runtime.GOOS, runtime.GOARCH)
			dest := filepath.Join(moduleDir, fmt.Sprintf("kusion-module-%s", dep.FullName))
			if runtime.GOOS == "windows" {
				source = fmt.Sprintf("%s.exe", source)
				dest = fmt.Sprintf("%s.exe", dest)
			}
			err = io.CopyFile(source, dest)
			allErrs = append(allErrs, err)
		}
	}

	if allErrs != nil {
		return utilerrors.NewAggregate(allErrs)
	}

	return nil
}
