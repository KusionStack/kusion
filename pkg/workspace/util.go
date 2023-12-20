package workspace

import (
	"errors"
	"fmt"
	"os"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

var ErrEmptyProjectName = errors.New("empty project name")

// CompleteWorkspace sets the workspace name and default value of unset item, should be called after ValidateWorkspace.
// The config items set as environment variables are not got by CompleteWorkspace.
func CompleteWorkspace(ws *workspace.Workspace, name string) {
	if ws.Name != "" {
		ws.Name = name
	}
	if ws.Backends != nil && GetBackendName(ws.Backends) == workspace.BackendMysql {
		CompleteMysqlConfig(ws.Backends.Mysql)
	}
}

// GetProjectModuleConfigs returns the module configs of a specified project, whose key is the module name,
// should be called after ValidateModuleConfigs.
// If got empty module configs, return nil config and nil error.
func GetProjectModuleConfigs(configs workspace.ModuleConfigs, projectName string) (map[string]workspace.GenericConfig, error) {
	if len(configs) == 0 {
		return nil, nil
	}
	if projectName == "" {
		return nil, ErrEmptyProjectName
	}

	projectCfgs := make(map[string]workspace.GenericConfig)
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
func GetProjectModuleConfig(config *workspace.ModuleConfig, projectName string) (workspace.GenericConfig, error) {
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
func getProjectModuleConfig(config *workspace.ModuleConfig, projectName string) (workspace.GenericConfig, error) {
	projectCfg := config.Default
	if len(projectCfg) == 0 {
		projectCfg = make(workspace.GenericConfig)
	}

	for name, cfg := range config.ModulePatcherConfigs {
		if name == workspace.DefaultBlock {
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

// GetKubernetesConfig returns kubernetes config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty kubernetes config, return nil.
func GetKubernetesConfig(configs *workspace.RuntimeConfigs) *workspace.KubernetesConfig {
	if configs == nil {
		return nil
	}
	return configs.Kubernetes
}

// GetTerraformConfig returns terraform config from runtime config, should be called after
// ValidateRuntimeConfigs.
// If got empty terraform config, return nil.
func GetTerraformConfig(configs *workspace.RuntimeConfigs) workspace.TerraformConfig {
	if configs == nil {
		return nil
	}
	return configs.Terraform
}

// GetProviderConfig returns the specified terraform provider config from runtime config, should be called
// after ValidateRuntimeConfigs.
// If got empty terraform config, return nil config and nil error.
func GetProviderConfig(configs *workspace.RuntimeConfigs, providerName string) (*workspace.ProviderConfig, error) {
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
func GetBackendName(configs *workspace.BackendConfigs) string {
	if configs == nil {
		return workspace.BackendLocal
	}
	if configs.Local != nil {
		return workspace.BackendLocal
	}
	if configs.Mysql != nil {
		return workspace.BackendMysql
	}
	if configs.Oss != nil {
		return workspace.BackendOss
	}
	if configs.S3 != nil {
		return workspace.BackendS3
	}
	return workspace.BackendLocal
}

// GetMysqlPasswordFromEnv returns mysql password set by environment variables.
func GetMysqlPasswordFromEnv() string {
	return os.Getenv(workspace.EnvBackendMysqlPassword)
}

// GetOssSensitiveDataFromEnv returns oss accessKeyID, accessKeySecret set by environment variables.
func GetOssSensitiveDataFromEnv() (string, string) {
	return os.Getenv(workspace.EnvOssAccessKeyID), os.Getenv(workspace.EnvOssAccessKeySecret)
}

// GetS3SensitiveDataFromEnv returns s3 accessKeyID, accessKeySecret, region set by environment variables.
func GetS3SensitiveDataFromEnv() (string, string, string) {
	region := os.Getenv(workspace.EnvAwsRegion)
	if region == "" {
		region = os.Getenv(workspace.EnvAwsDefaultRegion)
	}
	return os.Getenv(workspace.EnvAwsAccessKeyID), os.Getenv(workspace.EnvAwsSecretAccessKey), region
}

// CompleteMysqlConfig sets default value of mysql config if not set.
func CompleteMysqlConfig(config *workspace.MysqlConfig) {
	if config.Port == nil {
		port := workspace.DefaultMysqlPort
		config.Port = &port
	}
}

// CompleteWholeMysqlConfig constructs the whole mysql config by environment variables if set.
func CompleteWholeMysqlConfig(config *workspace.MysqlConfig) {
	password := GetMysqlPasswordFromEnv()
	if password != "" {
		config.Password = password
	}
}

// CompleteWholeOssConfig constructs the whole oss config by environment variables if set.
func CompleteWholeOssConfig(config *workspace.OssConfig) {
	accessKeyID, accessKeySecret := GetOssSensitiveDataFromEnv()
	if accessKeyID != "" {
		config.AccessKeyID = accessKeyID
	}
	if accessKeySecret != "" {
		config.AccessKeySecret = accessKeySecret
	}
}

// CompleteWholeS3Config constructs the whole s3 config by environment variables if set.
func CompleteWholeS3Config(config *workspace.S3Config) {
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

// GetIntFieldFromGenericConfig returns the value of the key in config which should be of type int.
// If exist but not int, return error. If not exist, return 0, nil.
func GetIntFieldFromGenericConfig(config workspace.GenericConfig, key string) (int, error) {
	value, ok := config[key]
	if !ok {
		return 0, nil
	}
	i, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("the value of %s is not int", key)
	}
	return i, nil
}

// GetStringFieldFromGenericConfig returns the value of the key in config which should be of type string.
// If exist but not string, return error; If not exist, return "", nil.
func GetStringFieldFromGenericConfig(config workspace.GenericConfig, key string) (string, error) {
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

// GetMapFieldFromGenericConfig returns the value of the key in config which should be of type map[string]any.
// If exist but not map[string]any, return error; If not exist, return nil, nil.
func GetMapFieldFromGenericConfig(config workspace.GenericConfig, key string) (map[string]any, error) {
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

// GetStringMapFieldFromGenericConfig returns the value of the key in config which should be of type map[string]string.
// If exist but not map[string]string, return error; If not exist, return nil, nil.
func GetStringMapFieldFromGenericConfig(config workspace.GenericConfig, key string) (map[string]string, error) {
	m, err := GetMapFieldFromGenericConfig(config, key)
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
