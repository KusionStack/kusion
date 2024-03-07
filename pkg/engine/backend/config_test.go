package backend

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
)

func TestNewConfig(t *testing.T) {
	mysqlPort := 3306
	testcases := []struct {
		name                     string
		success                  bool
		workDir                  string
		configs                  *v1.DeprecatedBackendConfigs
		opts                     *BackendOptions
		setEnvFunc, unSetEnvFunc func()
		expectedConfig           *StateStorageConfig
	}{
		{
			name:         "default config",
			success:      true,
			workDir:      "/test_project/test_stack",
			configs:      nil,
			opts:         &BackendOptions{},
			setEnvFunc:   nil,
			unSetEnvFunc: nil,
			expectedConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendLocal,
				Config: map[string]any{
					"path": "/test_project/test_stack/kusion_state.yaml",
				},
			},
		},
		{
			name:    "empty backend options",
			success: true,
			workDir: "/testProject/testStack",
			configs: &v1.DeprecatedBackendConfigs{
				Mysql: &v1.DeprecatedMysqlConfig{
					DBName:   "kusion_db",
					User:     "kusion",
					Password: "do_not_recommend",
					Host:     "127.0.0.1",
					Port:     &mysqlPort,
				},
			},
			opts: &BackendOptions{},
			setEnvFunc: func() {
				_ = os.Setenv(v1.DeprecatedEnvBackendMysqlPassword, "kusion_password")
			},
			unSetEnvFunc: func() {
				_ = os.Unsetenv(v1.DeprecatedEnvBackendMysqlPassword)
			},
			expectedConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendMysql,
				Config: map[string]any{
					"dbName":   "kusion_db",
					"user":     "kusion",
					"password": "kusion_password",
					"host":     "127.0.0.1",
					"port":     3306,
				},
			},
		},
		{
			name:    "backend options override",
			success: true,
			workDir: "/testProject/testStack",
			configs: &v1.DeprecatedBackendConfigs{
				Mysql: &v1.DeprecatedMysqlConfig{
					DBName: "kusion_db",
					User:   "kusion",
					Host:   "127.0.0.1",
					Port:   &mysqlPort,
				},
			},
			opts: &BackendOptions{
				Type:   v1.DeprecatedBackendS3,
				Config: []string{"region=ua-east-2", "bucket=kusion_bucket"},
			},
			setEnvFunc: func() {
				_ = os.Setenv(v1.DeprecatedEnvAwsRegion, "ua-east-1")
				_ = os.Setenv(v1.DeprecatedEnvAwsAccessKeyID, "aws_ak_id")
				_ = os.Setenv(v1.DeprecatedEnvAwsSecretAccessKey, "aws_ak_secret")
			},
			unSetEnvFunc: func() {
				_ = os.Unsetenv(v1.DeprecatedEnvAwsDefaultRegion)
				_ = os.Unsetenv(v1.DeprecatedEnvOssAccessKeyID)
				_ = os.Unsetenv(v1.DeprecatedEnvAwsSecretAccessKey)
			},
			expectedConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendS3,
				Config: map[string]any{
					"region":          "ua-east-2",
					"accessKeyID":     "aws_ak_id",
					"accessKeySecret": "aws_ak_secret",
					"bucket":          "kusion_bucket",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnvFunc != nil {
				tc.setEnvFunc()
			}
			config, err := NewConfig(tc.workDir, tc.configs, tc.opts)
			if tc.unSetEnvFunc != nil {
				tc.unSetEnvFunc()
			}
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, *tc.expectedConfig, *config)
		})
	}
}

func TestStateStorageConfig_NewStateStorage(t *testing.T) {
	testcases := []struct {
		name                 string
		success              bool
		config               *StateStorageConfig
		expectedStateStorage states.StateStorage
	}{
		{
			name:    "local state storage",
			success: true,
			config: &StateStorageConfig{
				Type: v1.DeprecatedBackendLocal,
				Config: map[string]any{
					"path": "/test_project/test_stack/kusion_state.yaml",
				},
			},
			expectedStateStorage: &local.FileSystemState{
				Path: "/test_project/test_stack/kusion_state.yaml",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stateStorage, err := tc.config.NewStateStorage()
			assert.Equal(t, tc.success, err == nil)
			assert.True(t, reflect.DeepEqual(tc.expectedStateStorage, stateStorage))
		})
	}
}

func TestMergeConfig(t *testing.T) {
	testcases := []struct {
		name                   string
		backendType            string
		config, overrideConfig *StateStorageConfig
		envConfig              map[string]any
		mergedConfig           *StateStorageConfig
	}{
		{
			name:        "empty override config",
			backendType: v1.DeprecatedBackendLocal,
			config: &StateStorageConfig{
				Type: v1.DeprecatedBackendLocal,
				Config: map[string]any{
					"path": "/test_project/test_stack/kusion_state.yaml",
				},
			},
			overrideConfig: nil,
			envConfig:      nil,
			mergedConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendLocal,
				Config: map[string]any{
					"path": "/test_project/test_stack/kusion_state.yaml",
				},
			},
		},
		{
			name:        "same type override config",
			backendType: v1.DeprecatedBackendMysql,
			config: &StateStorageConfig{
				Type: v1.DeprecatedBackendMysql,
				Config: map[string]any{
					"dbName": "kusion_db",
					"user":   "kusion",
					"host":   "127.0.0.1",
					"port":   3306,
				},
			},
			overrideConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendMysql,
				Config: map[string]any{
					"dbName": "new_kusion_db",
					"user":   "new_kusion",
				},
			},
			envConfig: map[string]any{
				"password": "new_kusion_password",
			},
			mergedConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendMysql,
				Config: map[string]any{
					"dbName":   "new_kusion_db",
					"user":     "new_kusion",
					"password": "new_kusion_password",
					"host":     "127.0.0.1",
					"port":     3306,
				},
			},
		},
		{
			name:        "different type override config",
			backendType: v1.DeprecatedBackendOss,
			config: &StateStorageConfig{
				Type: v1.DeprecatedBackendMysql,
				Config: map[string]any{
					"dbName": "kusion_db",
					"user":   "kusion",
					"host":   "127.0.0.1",
					"port":   3306,
				},
			},
			overrideConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendOss,
				Config: map[string]any{
					"endpoint":        "oss-cn-hangzhou.aliyuncs.com",
					"bucket":          "kusion_test",
					"accessKeyID":     "kusion_test",
					"accessKeySecret": "kusion_test",
				},
			},
			envConfig: map[string]any{
				"accessKeyID":     "kusion_test_env",
				"accessKeySecret": "kusion_test_env",
			},
			mergedConfig: &StateStorageConfig{
				Type: v1.DeprecatedBackendOss,
				Config: map[string]any{
					"endpoint":        "oss-cn-hangzhou.aliyuncs.com",
					"bucket":          "kusion_test",
					"accessKeyID":     "kusion_test",
					"accessKeySecret": "kusion_test",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			config := mergeConfig(tc.backendType, tc.config, tc.overrideConfig, tc.envConfig)
			assert.Equal(t, *tc.mergedConfig, *config)
		})
	}
}
