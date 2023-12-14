package workspace

import (
	"errors"
	"fmt"
	"os"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

var ErrEmptyProjectName = errors.New("empty project name")

// CompleteWorkspace sets the workspace name and default value of unset item, should be called after Validatev1.
// The config items set as environment variables are not got by Completev1.
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
// If got empty module configs, return nil config and nil error.
func GetProjectModuleConfigs(configs v1.ModuleConfigs, projectName string) (map[string]v1.GenericConfig, error) {
	if len(configs) == 0 {
		return nil, nil
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	projectCfgs := make(map[string]v1.GenericConfig)
	for name, cfg := range configs {
		projectCfg, err := getProjectModuleConfig(cfg, projectName)
		if projectCfg == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%w, module name: %s", err, name)
		}
		if len(projectCfg) != 0 {
			projectCfgs[name] = projectCfg
		}
	}

	return projectCfgs, nil
}

// GetProjectModuleConfig returns the module config of a specified project, should be called after
// ValidateModuleConfig.
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

	return projectCfg, nil
}

// GetKubernetesConfig returns kubernetes config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty kubernetes config, return nil.
func GetKubernetesConfig(configs *v1.RuntimeConfigs) *v1.KubernetesConfig {
	if configs == nil {
		return nil
	}
	return configs.Kubernetes
}

// GetTerraformConfig returns terraform config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty terraform config, return nil.
func GetTerraformConfig(configs *v1.RuntimeConfigs) v1.TerraformConfig {
	if configs == nil {
		return nil
	}
	return configs.Terraform
}

// GetProviderConfig returns the specified terraform provider config from runtime config, should be called
// after ValidateRuntimeConfigs.
// If got empty terraform config, return nil config and nil error.
func GetProviderConfig(configs *v1.RuntimeConfigs, providerName string) (*v1.ProviderConfig, error) {
	if providerName == "" {
		return nil, ErrEmptyTerraformProviderName
	}
	config := GetTerraformConfig(configs)
	if config == nil {
		return nil, nil
	}
	return config[providerName], nil
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

// GetIntFromGenericConfig returns the value of the key in config which should be of type int.
// If exist but not int, return error. If not exist, return 0, nil.
func GetIntFromGenericConfig(config v1.GenericConfig, key string) (int, error) {
	value, ok := config[key]
	if !ok {
		return 0, nil
	}
	i, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("the value of %s is not map", key)
	}
	return i, nil
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
	m, ok := value.(map[string]any)
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
