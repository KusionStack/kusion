package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
)

func TestValidateSetCurrentBackend(t *testing.T) {
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSetCurrentBackend(tc.config, "", tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateUnsetCurrentBackend(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
	}{
		{
			name:    "valid unsetting current backend",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
		},
		{
			name:    "invalid unsetting default backend",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: v1.DefaultBackendName,
					Backends: map[string]*v1.BackendConfig{
						v1.DefaultBackendName: {Type: v1.BackendTypeLocal},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUnsetCurrentBackend(tc.config, "")
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateSetBackendConfig(t *testing.T) {
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSetBackendConfig(nil, "", tc.val)
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

func TestValidateSetBackendType(t *testing.T) {
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
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{},
				},
			},
			key: "backends.dev.type",
			val: "s3",
		},
		{
			name:    "invalid backend type unsupported type",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{},
				},
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
			val: "s3",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSetBackendType(tc.config, tc.key, tc.val)
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

func TestValidateSetBackendConfigItems(t *testing.T) {
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
						"dev": {Type: v1.BackendTypeS3},
					},
				},
			},
			key: "backends.dev.configs",
			val: map[string]any{
				v1.BackendGenericOssBucket: "kusion",
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
				v1.BackendGenericOssBucket: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSetBackendConfigItems(tc.config, tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateUnsetBackendConfigItems(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		key     string
	}{
		{
			name:    "valid backend config items key",
			success: true,
			key:     "backends.dev.configs",
		},
		{
			name:    "invalid backend config items key unset default backend",
			success: false,
			key:     "backends.default.configs",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUnsetBackendConfigItems(nil, tc.key)
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
						"dev": {Type: v1.BackendTypeOss},
					},
				},
			},
			key:          "backends.dev.configs.bucket",
			backendTypes: []string{v1.BackendTypeOss},
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
			key:          "backends.dev.configs.bucket",
			backendTypes: []string{v1.BackendTypeOss},
		},
		{
			name:    "invalid backend config item conflict backend type",
			success: false,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
			key:          "backends.dev.configs.bucket",
			backendTypes: []string{v1.BackendTypeOss},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkBackendTypeForBackendItem(tc.config, tc.key, tc.backendTypes...)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}
