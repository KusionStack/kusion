package workspace

import (
	"errors"
	"fmt"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

var (
	ErrEmptyWorkspaceName = errors.New("empty workspace name")

	ErrEmptyModuleName                      = errors.New("empty module name")
	ErrEmptyModuleConfig                    = errors.New("empty module config")
	ErrEmptyModuleConfigBlockName           = errors.New("empty block name in module config")
	ErrEmptyModuleConfigBlock               = errors.New("empty block in module config")
	ErrEmptyModuleConfigProjectSelector     = errors.New("empty projectSelector in module config patcher block")
	ErrNotEmptyModuleConfigProjectSelector  = errors.New("not empty projectSelector in module config default block")
	ErrInvalidModuleConfigProjectSelector   = errors.New("invalid projectSelector in module config patcher block")
	ErrRepeatedModuleConfigSelectedProjects = errors.New("project should not repeat in one patcher block's projectSelector")
	ErrMultipleModuleConfigSelectedProjects = errors.New("a project cannot assign in more than one patcher block's projectSelector")

	ErrEmptyKubeConfig              = errors.New("empty kubeconfig")
	ErrEmptyTerraformProviderName   = errors.New("empty terraform provider name")
	ErrEmptyTerraformProviderConfig = errors.New("empty terraform provider config")

	ErrEmptyLocalFilePath = errors.New("empty local file path")
)

// ValidateWorkspace is used to validate the workspace.Workspace is valid or not.
func ValidateWorkspace(ws *workspace.Workspace) error {
	if ws.Name == "" {
		return ErrEmptyWorkspaceName
	}
	if len(ws.Modules) != 0 {
		if err := ValidateModuleConfigs(ws.Modules); err != nil {
			return err
		}
	}
	if ws.Runtimes != nil {
		if err := ValidateRuntimeConfigs(ws.Runtimes); err != nil {
			return err
		}
	}
	if ws.Backends != nil {
		if err := ValidateBackendConfigs(ws.Backends); err != nil {
			return err
		}
	}
	return nil
}

// ValidateModuleConfigs validates the workspace.ModuleConfigs is valid or not.
func ValidateModuleConfigs(configs workspace.ModuleConfigs) error {
	for name, cfg := range configs {
		if name == "" {
			return ErrEmptyModuleName
		}
		if len(cfg) == 0 {
			return fmt.Errorf("%w, module name: %s", ErrEmptyModuleConfig, name)
		}
		if err := ValidateModuleConfig(cfg); err != nil {
			return fmt.Errorf("%w, module name: %s", err, name)
		}
	}

	return nil
}

// ValidateModuleConfig is used to validate the workspace.ModuleConfig is valid or not.
func ValidateModuleConfig(config workspace.ModuleConfig) error {
	// allProjects is used to inspect if there are repeated projects in projectSelector
	// field or not.
	allProjects := make(map[string]string)
	for name, cfg := range config {
		switch name {
		case "":
			return ErrEmptyModuleConfigBlockName

		// default block must not be empty and not have field projectSelector
		case workspace.DefaultBlock:
			if len(cfg) == 0 {
				return fmt.Errorf("%w, block name: %s", ErrEmptyModuleConfigBlock, workspace.DefaultBlock)
			}
			if _, ok := cfg[workspace.ProjectSelectorField]; ok {
				return ErrNotEmptyModuleConfigProjectSelector
			}

		// patcher block must have field projectSelector, can be deserialized to string slice,
		// and there should be no repeated projects.
		default:
			unstructuredProjects, ok := cfg[workspace.ProjectSelectorField]
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

// ValidateRuntimeConfigs is used to validate the workspace.RuntimeConfigs is valid or not.
func ValidateRuntimeConfigs(configs *workspace.RuntimeConfigs) error {
	if configs.Kubernetes != nil {
		if err := ValidateKubernetesConfig(configs.Kubernetes); err != nil {
			return err
		}
	}
	if len(configs.Terraform) != 0 {
		if err := ValidateTerraformConfig(configs.Terraform); err != nil {
			return err
		}
	}
	return nil
}

// ValidateKubernetesConfig is used to validate the workspace.KubernetesConfig is valid or not.
func ValidateKubernetesConfig(config *workspace.KubernetesConfig) error {
	if config.KubeConfig == "" {
		return ErrEmptyKubeConfig
	}
	return nil
}

// ValidateTerraformConfig is used to validate the workspace.TerraformConfig is valid or not.
func ValidateTerraformConfig(config workspace.TerraformConfig) error {
	for name, cfg := range config {
		if name == "" {
			return ErrEmptyTerraformProviderName
		}
		if len(cfg) == 0 {
			return ErrEmptyTerraformProviderConfig
		}
	}
	return nil
}

// ValidateBackendConfigs is used to validate workspace.BackendConfigs is valid or not.
func ValidateBackendConfigs(configs *workspace.BackendConfigs) error {
	if configs.Local != nil {
		if err := ValidateLocalFileConfig(configs.Local); err != nil {
			return err
		}
	}
	return nil
}

// ValidateLocalFileConfig is used to validate workspace.LocalFileConfig is valid or not.
func ValidateLocalFileConfig(config *workspace.LocalFileConfig) error {
	if config.Path == "" {
		return ErrEmptyLocalFilePath
	}
	return nil
}
