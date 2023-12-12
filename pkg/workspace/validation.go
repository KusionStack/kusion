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

	ErrEmptyKubeConfig                   = errors.New("empty kubeconfig")
	ErrEmptyTerraformProviderName        = errors.New("empty terraform provider name")
	ErrEmptyTerraformProviderConfig      = errors.New("empty terraform provider config")
	ErrEmptyTerraformProviderSource      = errors.New("empty provider source")
	ErrEmptyTerraformProviderVersion     = errors.New("empty provider version")
	ErrEmptyTerraformProviderConfigKey   = errors.New("empty provider config key")
	ErrEmptyTerraformProviderConfigValue = errors.New("empty provider config value")

	ErrMultipleBackends     = errors.New("more than one backend configured")
	ErrEmptyMysqlDBName     = errors.New("empty db name")
	ErrEmptyMysqlUser       = errors.New("empty mysql db user")
	ErrEmptyMysqlHost       = errors.New("empty mysql host")
	ErrInvalidMysqlPort     = errors.New("mysql port must be between 1 and 65535")
	ErrEmptyBucket          = errors.New("empty bucket")
	ErrEmptyAccessKeyID     = errors.New("empty access key id")
	ErrEmptyAccessKeySecret = errors.New("empty access key secret")
	ErrEmptyOssEndpoint     = errors.New("empty oss endpoint")
	ErrEmptyS3Region        = errors.New("empty s3 region")
)

// ValidateWorkspace is used to validate the workspace get or set in the storage, and does not validate the
// config which can get from environment variables, such as access key id in backend configs.
func ValidateWorkspace(ws *workspace.Workspace) error {
	if ws.Name == "" {
		return ErrEmptyWorkspaceName
	}
	if ws.Modules != nil {
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

// ValidateModuleConfigs validates the moduleConfigs is valid or not.
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

// ValidateModuleConfig is used to validate the moduleConfig is valid or not.
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

// ValidateRuntimeConfigs is used to validate the runtimeConfigs is valid or not.
func ValidateRuntimeConfigs(configs *workspace.RuntimeConfigs) error {
	if configs.Kubernetes != nil {
		if err := ValidateKubernetesConfig(configs.Kubernetes); err != nil {
			return err
		}
	}
	if configs.Terraform != nil {
		if err := ValidateTerraformConfig(configs.Terraform); err != nil {
			return err
		}
	}
	return nil
}

// ValidateKubernetesConfig is used to validate the kubernetesConfig is valid or not.
func ValidateKubernetesConfig(config *workspace.KubernetesConfig) error {
	if config.KubeConfig == "" {
		return ErrEmptyKubeConfig
	}
	return nil
}

// ValidateTerraformConfig is used to validate the terraformConfig is valid or not.
func ValidateTerraformConfig(config workspace.TerraformConfig) error {
	for name, cfg := range config {
		if name == "" {
			return ErrEmptyTerraformProviderName
		}
		if cfg == nil {
			return fmt.Errorf("%w of provider %s", ErrEmptyTerraformProviderConfig, name)
		}
		if err := ValidateProviderConfig(cfg); err != nil {
			return fmt.Errorf("invalid terraform provider %s: %w", name, err)
		}
	}
	return nil
}

// ValidateProviderConfig is used to validate the providerConfig is valid or not.
func ValidateProviderConfig(config *workspace.ProviderConfig) error {
	if config.Source == "" {
		return ErrEmptyTerraformProviderSource
	}
	if config.Version == "" {
		return ErrEmptyTerraformProviderVersion
	}
	for k, v := range config.GenericConfig {
		if k == "" {
			return ErrEmptyTerraformProviderConfigKey
		}
		if v == nil {
			return fmt.Errorf("%w of field %s", ErrEmptyTerraformProviderConfigValue, k)
		}
	}
	return nil
}

// ValidateBackendConfigs is used to validate backendConfigs is valid or not, and does not validate the
// configs which can get from environment variables, such as access key id, etc.
func ValidateBackendConfigs(configs *workspace.BackendConfigs) error {
	if configureMoreThanOneBackend(configs) {
		return ErrMultipleBackends
	}

	// cause only one backend can be configured, hence the validity of the only one non-nil backend
	// represents the validity of the backend.
	if configs.Mysql != nil {
		return ValidateMysqlConfig(configs.Mysql)
	}
	if configs.Oss != nil {
		if err := ValidateGenericObjectStorageConfig(&configs.Oss.GenericObjectStorageConfig); err != nil {
			return fmt.Errorf("%w of %s", err, workspace.BackendOss)
		}
		return nil
	}
	if configs.S3 != nil {
		if err := ValidateGenericObjectStorageConfig(&configs.S3.GenericObjectStorageConfig); err != nil {
			return fmt.Errorf("%w of %s", err, workspace.BackendS3)
		}
		return nil
	}
	return nil
}

// configureMoreThanOneBackend checks whether there are more than one backend configured.
func configureMoreThanOneBackend(configs *workspace.BackendConfigs) bool {
	// configCondition returns: 1, if the backend configured or not; 2, if configured more than one backend.
	configCondition := func(configured bool, hasNewConfig bool) (bool, bool) {
		return configured || hasNewConfig, configured && hasNewConfig
	}

	var configured, moreThanOneConfig bool
	configured = configs.Local != nil
	configured, moreThanOneConfig = configCondition(configured, configs.Mysql != nil)
	if moreThanOneConfig {
		return moreThanOneConfig
	}
	configured, moreThanOneConfig = configCondition(configured, configs.Oss != nil)
	if moreThanOneConfig {
		return moreThanOneConfig
	}
	_, moreThanOneConfig = configCondition(configured, configs.S3 != nil)
	return moreThanOneConfig
}

// ValidateMysqlConfig is used to validate mysqlConfig is valid or not.
func ValidateMysqlConfig(config *workspace.MysqlConfig) error {
	if config.DBName == "" {
		return ErrEmptyMysqlDBName
	}
	if config.User == "" {
		return ErrEmptyMysqlUser
	}
	if config.Host == "" {
		return ErrEmptyMysqlHost
	}
	if config.Port != nil && (*config.Port < 1 || *config.Port > 65535) {
		return ErrInvalidMysqlPort
	}
	return nil
}

// ValidateGenericObjectStorageConfig is used to validate ossConfig and s3Config is valid or not, where the
// sensitive data items set as environment variables are not included.
func ValidateGenericObjectStorageConfig(config *workspace.GenericObjectStorageConfig) error {
	if config.Bucket == "" {
		return ErrEmptyBucket
	}
	return nil
}

// ValidateWholeOssConfig is used to validate ossConfig is valid or not, where all the items are included.
// If valid, the config contains all valid items to new an oss client.
func ValidateWholeOssConfig(config *workspace.OssConfig) error {
	if err := validateWholeGenericObjectStorageConfig(&config.GenericObjectStorageConfig); err != nil {
		return fmt.Errorf("%w of %s", err, workspace.BackendOss)
	}
	if config.Endpoint == "" {
		return ErrEmptyOssEndpoint
	}
	return nil
}

// ValidateWholeS3Config is used to validate s3Config is valid or not, where all the items are included.
// If valid, the config  contains all valid items to new a s3 client.
func ValidateWholeS3Config(config *workspace.S3Config) error {
	if err := validateWholeGenericObjectStorageConfig(&config.GenericObjectStorageConfig); err != nil {
		return fmt.Errorf("%w of %s", err, workspace.BackendS3)
	}
	if config.Region == "" {
		return ErrEmptyS3Region
	}
	return nil
}

func validateWholeGenericObjectStorageConfig(config *workspace.GenericObjectStorageConfig) error {
	if err := ValidateGenericObjectStorageConfig(config); err != nil {
		return err
	}
	if config.AccessKeyID == "" {
		return ErrEmptyAccessKeyID
	}
	if config.AccessKeySecret == "" {
		return ErrEmptyAccessKeySecret
	}
	return nil
}
