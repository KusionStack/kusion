package workspace

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	defaultBlock         = "default"
	projectSelectorField = "projectSelector"
)

var (
	ErrEmptyModuleName                      = errors.New("empty module name")
	ErrEmptyModuleConfig                    = errors.New("empty module config")
	ErrEmptyModuleConfigBlockName           = errors.New("empty block name in module config")
	ErrEmptyModuleConfigBlock               = errors.New("empty block in module config")
	ErrEmptyModuleConfigProjectSelector     = errors.New("empty projectSelector in module config patcher block")
	ErrNotEmptyModuleConfigProjectSelector  = errors.New("not empty projectSelector in module config default block")
	ErrInvalidModuleConfigProjectSelector   = errors.New("invalid projectSelector in module config patcher block")
	ErrRepeatedModuleConfigSelectedProjects = errors.New("project should not repeat in one patcher block's projectSelector")
	ErrMultipleModuleConfigSelectedProjects = errors.New("a project cannot assign in more than one patcher block's projectSelector")
)

var ErrEmptyQueryProjectName = errors.New("empty query project name")

// ModuleConfigs is a set of multiple ModuleConfig, whose key is the module name.
type ModuleConfigs map[string]ModuleConfig

// Validate validates the ModuleConfigs is valid or not.
func (m ModuleConfigs) Validate() error {
	for name, cfg := range m {
		if name == "" {
			return ErrEmptyModuleName
		}
		if len(cfg) == 0 {
			return fmt.Errorf("%w, module name: %s", ErrEmptyModuleConfig, name)
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("%w, module name: %s", err, name)
		}
	}

	return nil
}

// GetProjectModuleConfigs returns the module configs of a specified project, whose key is the module name.
func (m ModuleConfigs) GetProjectModuleConfigs(projectName string) (map[string]GenericConfig, error) {
	if projectName == "" {
		return nil, ErrEmptyQueryProjectName
	}

	projectCfgs := make(map[string]GenericConfig)
	for name, cfg := range m {
		projectCfg, err := cfg.getProjectModuleConfig(projectName)
		if err != nil {
			return nil, fmt.Errorf("%w, module name: %s", err, name)
		}
		if len(projectCfg) != 0 {
			projectCfgs[name] = projectCfg
		}
	}
	return projectCfgs, nil
}

// ModuleConfig is the config of a module, which contains a default and several patcher blocks.
//
// The default block's key is "default", and value is the module inputs. The patcher blocks' keys
// are the patcher names, which are just block identifiers without specific meaning, but must
// not be "default". Besides module inputs, patcher block's value also contains a field named
// "projectSelector", whose value is a slice containing the project names that use the patcher
// configs. A project can only be assigned in a patcher's "projectSelector" field, the assignment
// in multiple patchers is not allowed. For a project, if not specified in the patcher block's
// "projectSelector" field, the default config will be used.
type ModuleConfig map[string]GenericConfig

// Validate is used to validate the ModuleConfig is valid or not.
func (m ModuleConfig) Validate() error {
	// allProjects is used to inspect if there are repeated projects in projectSelector
	// field or not.
	allProjects := make(map[string]string)
	for name, cfg := range m {
		switch name {
		case "":
			return ErrEmptyModuleConfigBlockName

		// default block must not be empty and not have field projectSelector
		case defaultBlock:
			if len(cfg) == 0 {
				return fmt.Errorf("%w, block name: %s", ErrEmptyModuleConfigBlock, defaultBlock)
			}
			if _, ok := cfg[projectSelectorField]; ok {
				return ErrNotEmptyModuleConfigProjectSelector
			}

		// patcher block must have field projectSelector, can be deserialized to string slice,
		// and there should be no repeated projects.
		default:
			unstructuredProjects, ok := cfg[projectSelectorField]
			if !ok {
				return fmt.Errorf("%w, patcher block: %s", ErrEmptyModuleConfigProjectSelector, name)
			}
			if len(cfg) == 1 {
				return fmt.Errorf("%w, patcher block: %s", ErrEmptyModuleConfigBlock, name)
			}
			// the projectSelector filed should be deserialized to a string slice.
			projects, err := parseProjectsFromProjectSelector(unstructuredProjects)
			if err != nil {
				return fmt.Errorf("%w, patcher block: %s", err, name)
			}
			// a project cannot assign in more than one patcher block.
			for _, project := range projects {
				var patcherName string
				patcherName, ok = allProjects[project]
				if ok {
					if patcherName == name {
						return fmt.Errorf("%w, patcher block: %s", ErrRepeatedModuleConfigSelectedProjects, name)
					} else {
						return fmt.Errorf("%w, conflict patcher block: %s, %s", ErrMultipleModuleConfigSelectedProjects, name, patcherName)
					}
				}
				allProjects[project] = name
			}
		}
	}

	return nil
}

// GetProjectModuleConfig returns the module config of a specified project, should be called after Validate.
func (m ModuleConfig) GetProjectModuleConfig(projectName string) (GenericConfig, error) {
	if projectName == "" {
		return nil, ErrEmptyQueryProjectName
	}

	return m.getProjectModuleConfig(projectName)
}

// getProjectModuleConfig gets the module config of a specified project without checking the correctness
// of project name.
func (m ModuleConfig) getProjectModuleConfig(projectName string) (GenericConfig, error) {
	projectCfg := m[defaultBlock]
	if len(projectCfg) == 0 {
		projectCfg = make(GenericConfig)
	}

	for name, cfg := range m {
		if name == defaultBlock {
			continue
		}
		projects, err := parseProjectsFromProjectSelector(cfg[projectSelectorField])
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
				if k == projectSelectorField {
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
