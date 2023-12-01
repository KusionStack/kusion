package workspace

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

var (
	ErrEmptyQueryProjectName = errors.New("empty query project name")

	ErrEmptyQueryTerraformProviderName = errors.New("empty query terraform provider name")
	ErrNotExistTerraformProviderConfig = errors.New("not exist terraform provider config")
)

// GetProjectModuleConfigs returns the module configs of a specified project, whose key is the module name.
func GetProjectModuleConfigs(configs *workspace.ModuleConfigs, projectName string) (map[string]workspace.GenericConfig, error) {
	if projectName == "" {
		return nil, ErrEmptyQueryProjectName
	}

	projectCfgs := make(map[string]workspace.GenericConfig)
	for name, cfg := range *configs {
		projectCfg, err := getProjectModuleConfig(&cfg, projectName)
		if err != nil {
			return nil, fmt.Errorf("%w, module name: %s", err, name)
		}
		if len(projectCfg) != 0 {
			projectCfgs[name] = projectCfg
		}
	}
	return projectCfgs, nil
}

// GetProjectModuleConfig returns the module config of a specified project, should be called after Validate.
func GetProjectModuleConfig(config *workspace.ModuleConfig, projectName string) (workspace.GenericConfig, error) {
	if projectName == "" {
		return nil, ErrEmptyQueryProjectName
	}

	return getProjectModuleConfig(config, projectName)
}

// getProjectModuleConfig gets the module config of a specified project without checking the correctness
// of project name.
func getProjectModuleConfig(config *workspace.ModuleConfig, projectName string) (workspace.GenericConfig, error) {
	projectCfg := (*config)[workspace.DefaultBlock]
	if len(projectCfg) == 0 {
		projectCfg = make(workspace.GenericConfig)
	}

	for name, cfg := range *config {
		if name == workspace.DefaultBlock {
			continue
		}
		projects, err := parseProjectsFromProjectSelector(cfg[workspace.ProjectSelectorField])
		if err != nil {
			return nil, fmt.Errorf("%w, patcher block: %s", err, name)
		}
		// check the project is assigned in the block or not.
		var contain bool
		for _, project := range projects {
			if projectName == project {
				contain = true
				break
			}
		}
		if contain {
			for k, v := range cfg {
				if k == workspace.ProjectSelectorField {
					continue
				}
				projectCfg[k] = v
			}
			break
		}
	}

	return projectCfg, nil
}

// parseProjectsFromProjectSelector parses the projects in projectSelector field to string slice.
func parseProjectsFromProjectSelector(unstructuredProjects any) ([]string, error) {
	var projects []string
	bytes, err := yaml.Marshal(unstructuredProjects)
	if err != nil {
		return nil, fmt.Errorf("%w, marshal failed: %v", ErrInvalidModuleConfigProjectSelector, err)
	}
	if err = yaml.Unmarshal(bytes, &projects); err != nil {
		return nil, fmt.Errorf("%w, unmarshal failed: %v", ErrInvalidModuleConfigProjectSelector, err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("%w, empty projects", ErrInvalidModuleConfigProjectSelector)
	}
	return projects, nil
}

// GetTerraformProviderConfig is used to get a specified provider config.
func GetTerraformProviderConfig(config *workspace.TerraformConfig, providerName string) (workspace.GenericConfig, error) {
	if providerName == "" {
		return nil, ErrEmptyQueryTerraformProviderName
	}
	cfg, ok := (*config)[providerName]
	if !ok {
		return nil, ErrNotExistTerraformProviderConfig
	}
	return cfg, nil
}
