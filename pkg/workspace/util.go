package workspace

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

var (
	ErrEmptyProjectName             = errors.New("empty project name")
	ErrEmptyModuleConfigs           = errors.New("empty module configs")
	ErrNotExistProjectModuleConfigs = errors.New("not exist module configs of the project")
	ErrNotExistProjectModuleConfig  = errors.New("not exist module config of the project")

	ErrEmptyRuntimeConfigs             = errors.New("empty runtime configs")
	ErrNotExistKubernetesConfig        = errors.New("not exist kubernetes config")
	ErrNotExistTerraformConfig         = errors.New("not exist terraform config")
	ErrNotExistTerraformProviderConfig = errors.New("not exist terraform provider config")
)

// GetProjectModuleConfigs returns the module configs of a specified project, whose key is the module name,
// should be called after ValidateModuleConfigs.
// If got empty module configs, ErrNotExistProjectModuleConfigs will get returned.
func GetProjectModuleConfigs(configs workspace.ModuleConfigs, projectName string) (map[string]workspace.GenericConfig, error) {
	if len(configs) == 0 {
		return nil, ErrEmptyModuleConfigs
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	projectCfgs := make(map[string]workspace.GenericConfig)
	for name, cfg := range configs {
		projectCfg, err := getProjectModuleConfig(cfg, projectName)
		if errors.Is(err, ErrNotExistProjectModuleConfig) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%w, module name: %s", err, name)
		}
		if len(projectCfg) != 0 {
			projectCfgs[name] = projectCfg
		}
	}

	if len(projectCfgs) == 0 {
		return nil, ErrNotExistProjectModuleConfigs
	}
	return projectCfgs, nil
}

// GetProjectModuleConfig returns the module config of a specified project, should be called after
// ValidateModuleConfig.
// If got empty module config, ErrNotExistProjectModuleConfig will get returned.
func GetProjectModuleConfig(config workspace.ModuleConfig, projectName string) (workspace.GenericConfig, error) {
	if len(config) == 0 {
		return nil, ErrEmptyModuleConfig
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	return getProjectModuleConfig(config, projectName)
}

// getProjectModuleConfig gets the module config of a specified project without checking the correctness
// of project name.
func getProjectModuleConfig(config workspace.ModuleConfig, projectName string) (workspace.GenericConfig, error) {
	projectCfg := config[workspace.DefaultBlock]
	if len(projectCfg) == 0 {
		projectCfg = make(workspace.GenericConfig)
	}

	for name, cfg := range config {
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

	if len(projectCfg) == 0 {
		return nil, ErrNotExistProjectModuleConfig
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

// GetKubernetesConfig returns kubernetes config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty kubernetes config, ErrNotExistKubernetesConfig will get returned.
func GetKubernetesConfig(configs *workspace.RuntimeConfigs) (*workspace.KubernetesConfig, error) {
	if configs == nil {
		return nil, ErrEmptyRuntimeConfigs
	}
	if configs.Kubernetes == nil {
		return nil, ErrNotExistKubernetesConfig
	}
	return configs.Kubernetes, nil
}

// GetTerraformConfig returns terraform config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty terraform config, ErrNotExistTerraformConfig will get returned.
func GetTerraformConfig(configs *workspace.RuntimeConfigs) (workspace.TerraformConfig, error) {
	if configs == nil {
		return nil, ErrEmptyRuntimeConfigs
	}
	if len(configs.Terraform) == 0 {
		return nil, ErrNotExistTerraformConfig
	}
	return configs.Terraform, nil
}

// GetTerraformProviderConfig returns the specified terraform provider config from runtime config, should
// be called after ValidateRuntimeConfigs.
// If got empty terraform config, ErrNotExistTerraformProviderConfig will get returned.
func GetTerraformProviderConfig(configs *workspace.RuntimeConfigs, providerName string) (workspace.GenericConfig, error) {
	if providerName == "" {
		return nil, ErrEmptyTerraformProviderName
	}
	config, err := GetTerraformConfig(configs)
	if err != nil {
		return nil, err
	}

	cfg, ok := config[providerName]
	if !ok {
		return nil, ErrNotExistTerraformProviderConfig
	}
	return cfg, nil
}
