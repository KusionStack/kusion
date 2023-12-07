package workspace

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/apis/workspace"
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
// If got empty module configs, ErrEmptyProjectModuleConfigs will get returned.
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
		return nil, ErrEmptyProjectModuleConfig
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
// If got empty kubernetes config, ErrEmptyKubernetesConfig will get returned.
func GetKubernetesConfig(configs *workspace.RuntimeConfigs) (*workspace.KubernetesConfig, error) {
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
func GetTerraformConfig(configs *workspace.RuntimeConfigs) (workspace.TerraformConfig, error) {
	if configs == nil {
		return nil, ErrEmptyRuntimeConfigs
	}
	if len(configs.Terraform) == 0 {
		return nil, ErrEmptyTerraformConfig
	}
	return configs.Terraform, nil
}

// GetTerraformProviderConfig returns the specified terraform provider config from runtime config, should
// be called after ValidateRuntimeConfigs.
// If got empty terraform config, ErrEmptyTerraformProviderConfig will get returned.
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
		return nil, ErrEmptyTerraformProviderConfig
	}
	return cfg, nil
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
