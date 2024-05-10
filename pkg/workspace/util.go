package workspace

import (
	"errors"
	"fmt"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

var ErrEmptyProjectName = errors.New("empty project name")

// GetProjectModuleConfigs returns the module configs of a specified project, whose key is the module name, should be called after ValidateModuleConfigs.
// If got empty module configs, return nil config and nil error.
func GetProjectModuleConfigs(configs v1.ModuleConfigs, projectName string) (map[string]v1.GenericConfig, error) {
	if len(configs) == 0 {
		return nil, nil
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	projectConfigs := make(map[string]v1.GenericConfig)
	for name, cfg := range configs {
		moduleConfig, err := getProjectModuleConfig(cfg, projectName)
		if moduleConfig == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%w, module name: %s", err, name)
		}
		if len(moduleConfig) != 0 {
			projectConfigs[name] = moduleConfig
		}
	}

	return projectConfigs, nil
}

// GetProjectModuleConfig returns the module config of a specified project, should be called after ValidateModuleConfig.
// If got empty module config, return nil config and nil error.
func GetProjectModuleConfig(config *v1.ModuleConfig, projectName string) (v1.GenericConfig, error) {
	if config == nil {
		return nil, nil
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	return getProjectModuleConfig(config, projectName)
}

// getProjectModuleConfig gets the module config of a specified project without checking the correctness of project name.
func getProjectModuleConfig(config *v1.ModuleConfig, projectName string) (v1.GenericConfig, error) {
	projectCfg := config.Configs.Default
	if len(projectCfg) == 0 {
		projectCfg = make(v1.GenericConfig)
	}

	for name, cfg := range config.Configs.ModulePatcherConfigs {
		if name == v1.DefaultBlock {
			continue
		}
		// check the project is assigned in the block or not.
		var contain bool
		for _, project := range cfg.ProjectSelector {
			if projectName == project {
				contain = true
				break
			}
		}
		if contain {
			for k, v := range cfg.GenericConfig {
				if k == v1.ProjectSelectorField {
					continue
				}
				projectCfg[k] = v
			}
			break
		}
	}

	return projectCfg, nil
}

// GetInt32PointerFromGenericConfig returns the value of the key in config which should be of type int.
// If exist but not int, return error. If not exist, return nil.
func GetInt32PointerFromGenericConfig(config v1.GenericConfig, key string) (*int32, error) {
	value, ok := config[key]
	if !ok {
		return nil, nil
	}
	i, ok := value.(int)
	if !ok {
		return nil, fmt.Errorf("the value of %s is not int", key)
	}
	res := int32(i)
	return &res, nil
}

// GetStringFromGenericConfig returns the value of the key in config which should be of type string.
// If exist but not string, return error; If not exist, return "", nil.
func GetStringFromGenericConfig(config v1.GenericConfig, key string) (string, error) {
	value, ok := config[key]
	if !ok {
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("the value of %s is not string", key)
	}
	return s, nil
}

// GetMapFromGenericConfig returns the value of the key in config which should be of type map[string]any.
// If exist but not map[string]any, return error; If not exist, return nil, nil.
func GetMapFromGenericConfig(config v1.GenericConfig, key string) (map[string]any, error) {
	value, ok := config[key]
	if !ok {
		return nil, nil
	}
	m, ok := value.(v1.GenericConfig)
	if !ok {
		return nil, fmt.Errorf("the value of %s is not map", key)
	}
	return m, nil
}

// GetStringMapFromGenericConfig returns the value of the key in config which should be of type map[string]string.
// If exist but not map[string]string, return error; If not exist, return nil, nil.
func GetStringMapFromGenericConfig(config v1.GenericConfig, key string) (map[string]string, error) {
	m, err := GetMapFromGenericConfig(config, key)
	if err != nil {
		return nil, err
	}
	stringMap := make(map[string]string)
	for k, v := range m {
		stringValue, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("the value of %s.%s is not string", key, k)
		}
		stringMap[k] = stringValue
	}
	return stringMap, nil
}
