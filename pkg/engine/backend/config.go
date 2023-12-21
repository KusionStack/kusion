package backend

import (
	"fmt"
	"path"

	"github.com/zclconf/go-cty/cty/gocty"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	backendinit "kusionstack.io/kusion/pkg/engine/backend/init"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/workspace"
)

// StateStorageConfig contains backend config for state storage.
type StateStorageConfig struct {
	Type   string
	Config map[string]any
}

// NewConfig news a StateStorageConfig from workspace BackendConfigs, BackendOptions and environment variables.
func NewConfig(workDir string, configs *v1.BackendConfigs, opts *BackendOptions) (*StateStorageConfig, error) {
	var config, overrideConfig *StateStorageConfig
	config = convertWorkspaceBackendConfig(workDir, configs)
	if opts != nil && !opts.IsEmpty() {
		var err error
		if overrideConfig, err = opts.toStateStorageConfig(); err != nil {
			return nil, err
		}
	}

	backendType := config.Type
	if overrideConfig != nil && overrideConfig.Type != backendType {
		backendType = overrideConfig.Type
	}
	envConfig := getEnvBackendConfig(backendType)
	return mergeConfig(backendType, config, overrideConfig, envConfig), nil
}

// NewDefaultStateStorageConfig news the default state storage which uses local backend.
func NewDefaultStateStorageConfig(workDir string) *StateStorageConfig {
	return &StateStorageConfig{
		Type: v1.BackendLocal,
		Config: map[string]any{
			"path": path.Join(workDir, local.KusionStateFileFile),
		},
	}
}

// NewStateStorage news a StateStorage using the StateStorageConfig.
func (c *StateStorageConfig) NewStateStorage() (states.StateStorage, error) {
	backendFunc := backendinit.GetBackend(c.Type)
	if backendFunc == nil {
		return nil, fmt.Errorf("do not support state backend type %s", c.Type)
	}
	bf := backendFunc()
	backendSchema := bf.ConfigSchema()
	ctyBackend, err := gocty.ToCtyValue(c.Config, backendSchema)
	if err != nil {
		return nil, err
	}
	err = bf.Configure(ctyBackend)
	if err != nil {
		return nil, err
	}
	return bf.StateStorage(), nil
}

// convertWorkspaceBackendConfig converts workspace backend config to StateStorageConfig.
func convertWorkspaceBackendConfig(workDir string, configs *v1.BackendConfigs) *StateStorageConfig {
	name := workspace.GetBackendName(configs)
	var config map[string]any
	switch name {
	case v1.BackendLocal:
		config = NewDefaultStateStorageConfig(workDir).Config
	case v1.BackendMysql:
		config = map[string]any{
			"dbName":   configs.Mysql.DBName,
			"user":     configs.Mysql.User,
			"password": configs.Mysql.Password,
			"host":     configs.Mysql.Host,
			"port":     *configs.Mysql.Port,
		}
	case v1.BackendOss:
		config = map[string]any{
			"endpoint":        configs.Oss.Endpoint,
			"bucket":          configs.Oss.Bucket,
			"accessKeyID":     configs.Oss.AccessKeyID,
			"accessKeySecret": configs.Oss.AccessKeySecret,
		}
	case v1.BackendS3:
		config = map[string]any{
			"endpoint":        configs.S3.Endpoint,
			"bucket":          configs.S3.Bucket,
			"accessKeyID":     configs.S3.AccessKeyID,
			"accessKeySecret": configs.S3.AccessKeySecret,
			"region":          configs.S3.Region,
		}
	}
	return &StateStorageConfig{
		Type:   name,
		Config: config,
	}
}

// getEnvBackendConfig gets specified backend config set by environment variables
func getEnvBackendConfig(backendType string) map[string]any {
	config := make(map[string]any)
	switch backendType {
	case v1.BackendMysql:
		password := workspace.GetMysqlPasswordFromEnv()
		if password != "" {
			config["password"] = password
		}
	case v1.BackendOss:
		accessKeyID, accessKeySecret := workspace.GetOssSensitiveDataFromEnv()
		if accessKeyID != "" {
			config["accessKeyID"] = accessKeyID
		}
		if accessKeySecret != "" {
			config["accessKeySecret"] = accessKeySecret
		}
	case v1.BackendS3:
		accessKeyID, accessKeySecret, region := workspace.GetS3SensitiveDataFromEnv()
		if accessKeyID != "" {
			config["accessKeyID"] = accessKeyID
		}
		if accessKeySecret != "" {
			config["accessKeySecret"] = accessKeySecret
		}
		if region != "" {
			config["region"] = region
		}
	}
	return config
}

// mergeConfig merges the cli backend config (overrideConfig), environment variables (envConfig), and
// workspace backend config (config) in descending order of priority, to generate the StateStorageConfig
// which is used to new the StateStorage.
func mergeConfig(backendType string, config, overrideConfig *StateStorageConfig, envConfig map[string]any) *StateStorageConfig {
	var useConfig, useOverride bool
	if overrideConfig == nil {
		useConfig = true
	} else if overrideConfig.Type == config.Type {
		useConfig = true
		useOverride = true
	} else {
		useOverride = true
	}

	mergedConfig := &StateStorageConfig{
		Type:   backendType,
		Config: make(map[string]any),
	}
	if useConfig {
		for k, v := range config.Config {
			mergedConfig.Config[k] = v
		}
	}
	for k, v := range envConfig {
		mergedConfig.Config[k] = v
	}
	if useOverride {
		for k, v := range overrideConfig.Config {
			mergedConfig.Config[k] = v
		}
	}
	return mergedConfig
}
