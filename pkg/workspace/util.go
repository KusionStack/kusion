package workspace

import (
	"errors"
	"fmt"
	"os"

	"kusionstack.io/kusion/pkg/apis/core/v1"
)

var (
	ErrEmptyProjectName          = errors.New("empty project name")
	ErrEmptyModuleConfigs        = errors.New("empty module configs")
	ErrEmptyProjectModuleConfigs = errors.New("empty module configs of the project")
	ErrEmptyProjectModuleConfig  = errors.New("empty module config of the project")

	ErrEmptyRuntimeConfigs   = errors.New("empty runtime configs")
	ErrEmptyKubernetesConfig = errors.New("empty kubernetes config")
	ErrEmptyTerraformConfig  = errors.New("empty terraform config")
)

// CompleteWorkspace sets the workspace name and default value of unset item, should be called after ValidateWorkspace.
// The config items set as environment variables are not got by CompleteWorkspace.
func CompleteWorkspace(ws *v1.Workspace, name string) {
	if ws.Name != "" {
		ws.Name = name
	}
	if ws.Backends != nil && GetBackendName(ws.Backends) == v1.BackendMysql {
		CompleteMysqlConfig(ws.Backends.Mysql)
	}
}

// GetProjectModuleConfigs returns the module configs of a specified project, whose key is the module name,
// should be called after ValidateModuleConfigs.
// If got empty module configs, ErrEmptyProjectModuleConfigs will get returned.
func GetProjectModuleConfigs(configs v1.ModuleConfigs, projectName string) (map[string]v1.GenericConfig, error) {
	if len(configs) == 0 {
		return nil, ErrEmptyModuleConfigs
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	projectCfgs := make(map[string]v1.GenericConfig)
	for name, cfg := range configs {
		projectCfg, err := getProjectModuleConfig(cfg, projectName)
		if errors.Is(err, ErrEmptyProjectModuleConfig) {
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
		return nil, ErrEmptyProjectModuleConfigs
	}
	return projectCfgs, nil
}

// GetProjectModuleConfig returns the module config of a specified project, should be called after
// ValidateModuleConfig.
// If got empty module config, ErrEmptyProjectModuleConfig will get returned.
func GetProjectModuleConfig(config *v1.ModuleConfig, projectName string) (v1.GenericConfig, error) {
	if config == nil {
		return nil, ErrEmptyModuleConfig
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	return getProjectModuleConfig(config, projectName)
}

// getProjectModuleConfig gets the module config of a specified project without checking the correctness
// of project name.
func getProjectModuleConfig(config *v1.ModuleConfig, projectName string) (v1.GenericConfig, error) {
	projectCfg := config.Default
	if len(projectCfg) == 0 {
		projectCfg = make(v1.GenericConfig)
	}

	for name, cfg := range config.ModulePatcherConfigs {
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

	if len(projectCfg) == 0 {
		return nil, ErrEmptyProjectModuleConfig
	}
	return projectCfg, nil
}

// GetKubernetesConfig returns kubernetes config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty kubernetes config, ErrEmptyKubernetesConfig will get returned.
func GetKubernetesConfig(configs *v1.RuntimeConfigs) (*v1.KubernetesConfig, error) {
	if configs == nil {
		return nil, ErrEmptyRuntimeConfigs
	}
	if configs.Kubernetes == nil {
		return nil, ErrEmptyKubernetesConfig
	}
	return configs.Kubernetes, nil
}

// GetTerraformConfig returns terraform config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty terraform config, ErrEmptyTerraformConfig will get returned.
func GetTerraformConfig(configs *v1.RuntimeConfigs) (v1.TerraformConfig, error) {
	if configs == nil {
		return nil, ErrEmptyRuntimeConfigs
	}
	if len(configs.Terraform) == 0 {
		return nil, ErrEmptyTerraformConfig
	}
	return configs.Terraform, nil
}

// GetProviderConfig returns the specified terraform provider config from runtime config, should be called
// after ValidateRuntimeConfigs.
// If got empty terraform config, ErrEmptyTerraformProviderConfig will get returned.
func GetProviderConfig(configs *v1.RuntimeConfigs, providerName string) (*v1.ProviderConfig, error) {
	if providerName == "" {
		return nil, ErrEmptyTerraformProviderName
	}
	config, err := GetTerraformConfig(configs)
	if err != nil {
		return nil, err
	}

	cfg, ok := config[providerName]
	if !ok {
		return nil, ErrEmptyTerraformProviderConfig
	}
	return cfg, nil
}

// GetBackendName returns the backend name that is configured in BackendConfigs, should be called after
// ValidateBackendConfigs.
func GetBackendName(configs *v1.BackendConfigs) string {
	if configs == nil {
		return v1.BackendLocal
	}
	if configs.Local != nil {
		return v1.BackendLocal
	}
	if configs.Mysql != nil {
		return v1.BackendMysql
	}
	if configs.Oss != nil {
		return v1.BackendOss
	}
	if configs.S3 != nil {
		return v1.BackendS3
	}
	return v1.BackendLocal
}

// GetMysqlPasswordFromEnv returns mysql password set by environment variables.
func GetMysqlPasswordFromEnv() string {
	return os.Getenv(v1.EnvBackendMysqlPassword)
}

// GetOssSensitiveDataFromEnv returns oss accessKeyID, accessKeySecret set by environment variables.
func GetOssSensitiveDataFromEnv() (string, string) {
	return os.Getenv(v1.EnvOssAccessKeyID), os.Getenv(v1.EnvOssAccessKeySecret)
}

// GetS3SensitiveDataFromEnv returns s3 accessKeyID, accessKeySecret, region set by environment variables.
func GetS3SensitiveDataFromEnv() (string, string, string) {
	region := os.Getenv(v1.EnvAwsRegion)
	if region == "" {
		region = os.Getenv(v1.EnvAwsDefaultRegion)
	}
	return os.Getenv(v1.EnvAwsAccessKeyID), os.Getenv(v1.EnvAwsSecretAccessKey), region
}

// CompleteMysqlConfig sets default value of mysql config if not set.
func CompleteMysqlConfig(config *v1.MysqlConfig) {
	if config.Port == nil {
		port := v1.DefaultMysqlPort
		config.Port = &port
	}
}

// CompleteWholeMysqlConfig constructs the whole mysql config by environment variables if set.
func CompleteWholeMysqlConfig(config *v1.MysqlConfig) {
	password := GetMysqlPasswordFromEnv()
	if password != "" {
		config.Password = password
	}
}

// CompleteWholeOssConfig constructs the whole oss config by environment variables if set.
func CompleteWholeOssConfig(config *v1.OssConfig) {
	accessKeyID, accessKeySecret := GetOssSensitiveDataFromEnv()
	if accessKeyID != "" {
		config.AccessKeyID = accessKeyID
	}
	if accessKeySecret != "" {
		config.AccessKeySecret = accessKeySecret
	}
}

// CompleteWholeS3Config constructs the whole s3 config by environment variables if set.
func CompleteWholeS3Config(config *v1.S3Config) {
	accessKeyID, accessKeySecret, region := GetS3SensitiveDataFromEnv()
	if accessKeyID != "" {
		config.AccessKeyID = accessKeyID
	}
	if accessKeySecret != "" {
		config.AccessKeySecret = accessKeySecret
	}
	if region != "" {
		config.Region = region
	}
}
