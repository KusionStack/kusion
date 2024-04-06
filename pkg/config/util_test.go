package config

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
)

func mockNewOperator(config *v1.Config) {
	mockey.Mock(newOperator).Return(&operator{
		configFilePath:  mockConfigPath,
		registeredItems: newRegisteredItems(),
		config:          config,
	}, nil).Build()
}

func TestGetConfig(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		expectedConfig *v1.Config
	}{
		{
			name:           "get config successfully",
			success:        true,
			expectedConfig: mockValidConfig(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.expectedConfig)
				mockey.Mock((*operator).readConfig).Return(nil)
				config, err := GetConfig()
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedConfig, config)
			})
		})
	}
}

func TestGetEncodedConfigItem(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		config             *v1.Config
		configItemKey      string
		expectedConfigItem string
	}{
		{
			name:               "get encoded config item successfully type string",
			success:            true,
			config:             mockValidConfig(),
			configItemKey:      "backends.dev.type",
			expectedConfigItem: "local",
		},
		{
			name:               "get encoded config item successfully type int",
			success:            true,
			config:             mockValidConfig(),
			configItemKey:      "backends.pre.configs.port",
			expectedConfigItem: "3306",
		},
		{
			name:               "get encoded config item successfully type struct",
			success:            true,
			config:             mockValidConfig(),
			configItemKey:      "backends.prod",
			expectedConfigItem: `{"configs":{"bucket":"kusion"},"type":"s3"}`,
		},
		{
			name:               "failed to get encoded config item empty item",
			success:            false,
			config:             nil,
			configItemKey:      "backends.stage",
			expectedConfigItem: "",
		},
		{
			name:               "failed to get encoded config item not registered item",
			success:            false,
			config:             nil,
			configItemKey:      "backends.current.not.registered",
			expectedConfigItem: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.config)
				mockey.Mock((*operator).readConfig).Return(nil).Build()
				item, err := GetEncodedConfigItem(tc.configItemKey)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedConfigItem, item)
			})
		})
	}
}

func TestSetEncodedConfigItem(t *testing.T) {
	testcases := []struct {
		name                   string
		success                bool
		configItemKey          string
		configItem             string
		config, expectedConfig *v1.Config
	}{
		{
			name:          "set encoded config item successfully type string",
			success:       true,
			configItemKey: "backends.dev.type",
			configItem:    "local",
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{},
				},
			},
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: "local"},
					},
				},
			},
		},
		{
			name:          "set encoded config item successfully type struct",
			success:       true,
			configItemKey: "backends.pre",
			configItem:    `{"configs":{"dbName":"kusion","host":"127.0.0.1","port":3306,"user":"kk"},"type":"mysql"}`,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{},
				},
			},
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kusion",
								v1.BackendMysqlUser:   "kk",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   3306,
							},
						},
					},
				},
			},
		},
		{
			name:           "failed to set encoded config item empty key",
			success:        false,
			configItemKey:  "",
			configItem:     "dev",
			config:         nil,
			expectedConfig: nil,
		},
		{
			name:           "failed to set encoded config item empty value",
			success:        false,
			configItemKey:  "backends.dev.type",
			configItem:     "",
			config:         nil,
			expectedConfig: nil,
		},
		{
			name:           "failed to set encoded config item invalid value",
			success:        false,
			configItemKey:  "backends.dev.configs.port",
			configItem:     "-1",
			config:         mockValidConfig(),
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.config)
				mockey.Mock((*operator).writeConfig).Return(nil).Build()
				err := SetEncodedConfigItem(tc.configItemKey, tc.configItem)
				assert.Equal(t, tc.success, err == nil)
				if err == nil {
					var config *v1.Config
					config, err = GetConfig()
					assert.NoError(t, err)
					assert.Equal(t, tc.expectedConfig, config)
				}
			})
		})
	}
}

func TestDeleteConfigItem(t *testing.T) {
	testcases := []struct {
		name                   string
		success                bool
		configItemKey          string
		config, expectedConfig *v1.Config
	}{
		{
			name:          "delete config item successfully",
			success:       true,
			configItemKey: "backends.dev.type",
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						v1.DefaultBackendName: {Type: "local"},
						"dev":                 {Type: "local"},
					},
				},
			},
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						v1.DefaultBackendName: {Type: "local"},
					},
				},
			},
		},
		{
			name:          "failed to delete config item invalid unset",
			success:       false,
			configItemKey: "backends.pre.type",
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kusion",
								v1.BackendMysqlUser:   "kk",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   3306,
							},
						},
					},
				},
			},
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.config)
				mockey.Mock((*operator).writeConfig).Return(nil).Build()
				err := DeleteConfigItem(tc.configItemKey)
				assert.Equal(t, tc.success, err == nil)
				if err == nil {
					var config *v1.Config
					config, err = GetConfig()
					assert.NoError(t, err)
					assert.Equal(t, tc.expectedConfig, config)
				}
			})
		})
	}
}
