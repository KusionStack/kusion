package config

import (
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func mockNewOperator(configFilePath string, config *v1.Config) {
	mockey.Mock(newOperator).Return(&operator{
		configFilePath:  configFilePath,
		registeredItems: newRegisteredItems(),
		config:          config,
	}, nil).Build()
}

func TestGetConfig(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		configFilePath string
		expectedConfig *v1.Config
	}{
		{
			name:           "get config successfully",
			success:        true,
			configFilePath: existValidConfigPath,
			expectedConfig: validConfig,
		},
		{
			name:           "failed to get config empty config",
			success:        false,
			configFilePath: emptyValidConfigPath,
			expectedConfig: nil,
		},
		{
			name:           "failed to get config invalid config",
			success:        false,
			configFilePath: invalidConfigPath,
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.configFilePath, nil)
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
		configFilePath     string
		configItemKey      string
		expectedConfigItem string
	}{
		{
			name:               "get encoded config item successfully type string",
			success:            true,
			configFilePath:     existValidConfigPath,
			configItemKey:      "backends.dev.type",
			expectedConfigItem: "local",
		},
		{
			name:               "get encoded config item successfully type int",
			success:            true,
			configFilePath:     existValidConfigPath,
			configItemKey:      "backends.pre.configs.port",
			expectedConfigItem: "6443",
		},
		{
			name:               "get encoded config item successfully type struct",
			success:            true,
			configFilePath:     existValidConfigPath,
			configItemKey:      "backends.prod",
			expectedConfigItem: `{"configs":{"bucket":"kusion"},"type":"s3"}`,
		},
		{
			name:               "failed to get encoded config item empty item",
			success:            false,
			configFilePath:     emptyValidConfigPath,
			configItemKey:      "backends.stage",
			expectedConfigItem: "",
		},
		{
			name:               "failed to get encoded config item not registered item",
			success:            false,
			configFilePath:     emptyValidConfigPath,
			configItemKey:      "backends.current.not.registered",
			expectedConfigItem: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.configFilePath, nil)
				item, err := GetEncodedConfigItem(tc.configItemKey)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedConfigItem, item)
			})
		})
	}
}

func TestSetEncodedConfigItem(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		configFilePath string
		configItemKey  string
		configItem     string
		expectedConfig *v1.Config
	}{
		{
			name:           "set encoded config item successfully type string",
			success:        true,
			configFilePath: emptyValidConfigPath,
			configItemKey:  "backends.dev.type",
			configItem:     "local",
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: "local"},
					},
				},
			},
		},
		{
			name:           "set encoded config item successfully type struct",
			success:        true,
			configFilePath: emptyValidConfigPath,
			configItemKey:  "backends.pre",
			configItem:     `{"configs":{"dbName":"kk","host":"127.0.0.1","port":6443,"user":"kusion"},"type":"mysql"}`,
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kk",
								v1.BackendMysqlUser:   "kusion",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   6443,
							},
						},
					},
				},
			},
		},
		{
			name:           "failed to set encoded config item empty key",
			success:        false,
			configFilePath: emptyValidConfigPath,
			configItemKey:  "",
			configItem:     "dev",
			expectedConfig: nil,
		},
		{
			name:           "failed to set encoded config item empty value",
			success:        false,
			configFilePath: emptyValidConfigPath,
			configItemKey:  "backends.dev.type",
			configItem:     "",
			expectedConfig: nil,
		},
		{
			name:           "failed to set encoded config item invalid value",
			success:        false,
			configFilePath: existValidConfigPath,
			configItemKey:  "backends.dev.configs.port",
			configItem:     "-1",
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock config operator", t, func() {
				mockNewOperator(tc.configFilePath, nil)
				err := SetEncodedConfigItem(tc.configItemKey, tc.configItem)
				assert.Equal(t, tc.success, err == nil)
				if err == nil {
					var config *v1.Config
					config, err = GetConfig()
					assert.NoError(t, err)
					assert.Equal(t, tc.expectedConfig, config)
				}
				if tc.configFilePath == emptyValidConfigPath {
					_ = os.Remove(emptyValidConfigPath)
				}
			})
		})
	}
}

func TestDeleteConfigItem(t *testing.T) {
	testcases := []struct {
		name                   string
		success                bool
		configFilePath         string
		configItemKey          string
		config, expectedConfig *v1.Config
	}{
		{
			name:           "delete config item successfully",
			success:        true,
			configFilePath: emptyValidConfigPath,
			configItemKey:  "backends.dev.type",
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: "local"},
					},
				},
			},
			expectedConfig: nil,
		},
		{
			name:           "failed to delete config item invalid unset",
			success:        false,
			configFilePath: emptyValidConfigPath,
			configItemKey:  "backends.pre.type",
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kk",
								v1.BackendMysqlUser:   "kusion",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   6443,
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
				mockNewOperator(tc.configFilePath, tc.config)
				err := DeleteConfigItem(tc.configItemKey)
				assert.Equal(t, tc.success, err == nil)
				if err == nil {
					var config *v1.Config
					config, err = GetConfig()
					if tc.expectedConfig == nil {
						assert.Equal(t, ErrEmptyConfig, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tc.expectedConfig, config)
					}
				}
				if tc.configFilePath == emptyValidConfigPath {
					_ = os.Remove(emptyValidConfigPath)
				}
			})
		})
	}
}
