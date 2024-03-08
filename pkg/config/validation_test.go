package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func TestValidateCurrentBackend(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		val     string
	}{
		{
			name:    "valid current backend",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			val: "dev",
		},
		{
			name:    "invalid current backend not exist backend",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			val: "dev-not-exist",
		},
		{
			name:    "invalid current backend not exist config",
			success: false,
			config:  nil,
			val:     "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCurrentBackend(tc.config, "", tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateBackendConfig(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		val     *v1.BackendConfig
	}{
		{
			name:    "valid local backend",
			success: true,
			val: &v1.BackendConfig{
				Type: v1.BackendTypeLocal,
				Configs: map[string]any{
					v1.BackendLocalPath: "/etc",
				},
			},
		},
		{
			name:    "valid database backend",
			success: true,
			val: &v1.BackendConfig{
				Type: v1.BackendTypeMysql,
				Configs: map[string]any{
					v1.BackendMysqlDBName: "kusion",
					v1.BackendMysqlUser:   "kk",
					v1.BackendMysqlHost:   "127.0.0.1",
					v1.BackendMysqlPort:   3306,
				},
			},
		},
		{
			name:    "valid oss backend",
			success: true,
			val: &v1.BackendConfig{
				Type: v1.BackendTypeOss,
				Configs: map[string]any{
					v1.BackendGenericOssBucket:   "kusion",
					v1.BackendGenericOssEndpoint: "http://oss-cn-hangzhou.aliyuncs.com",
				},
			},
		},
		{
			name:    "valid s3 backend",
			success: true,
			val: &v1.BackendConfig{
				Type: v1.BackendTypeS3,
				Configs: map[string]any{
					v1.BackendGenericOssBucket:   "kusion",
					v1.BackendGenericOssEndpoint: "http://oss-cn-hangzhou.aliyuncs.com",
				},
			},
		},
		{
			name:    "invalid backend config invalid backend type",
			success: false,
			val: &v1.BackendConfig{
				Type: "not-support-type",
			},
		},
		{
			name:    "invalid backend config invalid config item type",
			success: false,
			val: &v1.BackendConfig{
				Type: v1.DeprecatedBackendMysql,
				Configs: map[string]any{
					v1.BackendMysqlDBName: "kusion",
					v1.BackendMysqlUser:   "kk",
					v1.BackendMysqlHost:   "127.0.0.1",
					v1.BackendMysqlPort:   "3306",
				},
			},
		},
		{
			name:    "invalid backend config unsupported config item",
			success: false,
			val: &v1.BackendConfig{
				Type: v1.DeprecatedBackendMysql,
				Configs: map[string]any{
					"not-support": "mock-not-support-value",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBackendConfig(nil, "", tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateUnsetBackendConfig(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		key     string
	}{
		{
			name:    "valid unset backend config",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			key: "backends.dev",
		},
		{
			name:    "invalid unset backend config in-use backend",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			key: "backends.dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUnsetBackendConfig(tc.config, tc.key)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateBackendType(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		key     string
		val     string
	}{
		{
			name:    "valid backend type new backend",
			success: true,
			config: &v1.Config{
				Backends: nil,
			},
			key: "backends.dev.type",
			val: "mysql",
		},
		{
			name:    "invalid backend type unsupported type",
			success: false,
			config: &v1.Config{
				Backends: nil,
			},
			key: "backends.dev.type",
			val: "not-supported",
		},
		{
			name:    "invalid backend type conflict assign",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type:    v1.BackendTypeLocal,
							Configs: map[string]any{v1.BackendLocalPath: "/etc"},
						},
					},
				},
			},
			key: "backends.dev.type",
			val: "mysql",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBackendType(tc.config, tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateUnsetBackendType(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		key     string
	}{
		{
			name:    "valid unset backend type",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			key: "backends.dev.type",
		},
		{
			name:    "invalid unset backend type non-empty config",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type:    v1.BackendTypeLocal,
							Configs: map[string]any{v1.BackendLocalPath: "/etc"},
						},
					},
				},
			},
			key: "backends.dev.type",
		},
		{
			name:    "invalid unset backend type in-use current",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			key: "backends.dev.type",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUnsetBackendType(tc.config, tc.key)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateBackendConfigItems(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		key     string
		val     map[string]any
	}{
		{
			name:    "valid backend config items",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeMysql},
					},
				},
			},
			key: "backends.dev.configs",
			val: map[string]any{
				v1.BackendMysqlDBName: "kusion",
				v1.BackendMysqlUser:   "kk",
				v1.BackendMysqlHost:   "127.0.0.1",
				v1.BackendMysqlPort:   3306,
			},
		},
		{
			name:    "invalid backend config items empty backend type",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {},
					},
				},
			},
			key: "backends.dev.configs",
			val: map[string]any{
				v1.BackendMysqlDBName: "kusion",
				v1.BackendMysqlUser:   "kk",
				v1.BackendMysqlHost:   "127.0.0.1",
				v1.BackendMysqlPort:   3306,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBackendConfigItems(tc.config, tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateMysqlBackendPort(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		key     string
		val     int
	}{
		{
			name:    "valid mysql port",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeMysql},
					},
				},
			},
			key: "backends.dev.configs.port",
			val: 3306,
		},
		{
			name:    "invalid mysql port",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeMysql},
					},
				},
			},
			key: "backends.dev.configs.port",
			val: -1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMysqlBackendPort(tc.config, tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestCheckBackendTypeForBackendItem(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		config       *v1.Config
		key          string
		backendTypes []string
	}{
		{
			name:    "valid backend config item",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeMysql},
					},
				},
			},
			key:          "backends.dev.configs.dbName",
			backendTypes: []string{v1.BackendTypeMysql},
		},
		{
			name:    "invalid backend config item empty backend type",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {},
					},
				},
			},
			key:          "backends.dev.configs.dbName",
			backendTypes: []string{v1.BackendTypeMysql},
		},
		{
			name:    "invalid backend config item conflict backend type",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeOss},
					},
				},
			},
			key:          "backends.dev.configs.dbName",
			backendTypes: []string{v1.BackendTypeMysql},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkBackendTypeForBackendItem(tc.config, tc.key, tc.backendTypes...)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}
