package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

var (
	testDataPath         = "testdata"
	existValidConfigPath = filepath.Join(testDataPath, "config_for_read.yaml")
	emptyValidConfigPath = filepath.Join(testDataPath, "config_for_write.yaml")
	invalidConfigPath    = filepath.Join(testDataPath, "config_invalid.yaml")

	registeredItems = newRegisteredItems()

	validConfig = &v1.Config{
		Backends: &v1.BackendConfigs{
			Current: "dev",
			Backends: map[string]*v1.BackendConfig{
				"dev": {
					Type: v1.BackendTypeLocal,
				},
				"pre": {
					Type: v1.BackendTypeMysql,
					Configs: map[string]any{
						v1.BackendMysqlDBName: "kusion",
						v1.BackendMysqlUser:   "kk",
						v1.BackendMysqlHost:   "127.0.0.1",
						v1.BackendMysqlPort:   3306,
					},
				},
				"prod": {
					Type: v1.BackendTypeS3,
					Configs: map[string]any{
						v1.BackendGenericOssBucket: "kusion",
					},
				},
			},
		},
	}
	validCfg = map[string]any{
		v1.ConfigBackends: map[string]any{
			v1.BackendCurrent: "dev",
			"dev": map[string]any{
				v1.BackendType: v1.BackendTypeLocal,
			},
			"pre": map[string]any{
				v1.BackendType: v1.DeprecatedBackendMysql,
				v1.BackendConfigItems: map[string]any{
					v1.BackendMysqlDBName: "kusion",
					v1.BackendMysqlUser:   "kk",
					v1.BackendMysqlHost:   "127.0.0.1",
					v1.BackendMysqlPort:   3306,
				},
			},
			"prod": map[string]any{
				v1.BackendType: v1.BackendTypeS3,
				v1.BackendConfigItems: map[string]any{
					v1.BackendGenericOssBucket: "kusion",
				},
			},
		},
	}
)

func mockOperator(configFilePath string, config *v1.Config) *operator {
	return &operator{
		configFilePath:  configFilePath,
		registeredItems: registeredItems,
		config:          config,
	}
}

func TestOperator_ReadConfig(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		o              *operator
		expectedConfig *v1.Config
	}{
		{
			name:           "read config successfully",
			success:        true,
			o:              mockOperator(existValidConfigPath, nil),
			expectedConfig: validConfig,
		},
		{
			name:           "read not exist config successfully",
			success:        true,
			o:              mockOperator(emptyValidConfigPath, nil),
			expectedConfig: nil,
		},
		{
			name:           "failed to read config invalid structure",
			success:        false,
			o:              mockOperator(invalidConfigPath, nil),
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.o.readConfig()
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedConfig, tc.o.config)
		})
	}
}

func TestOperator_WriteConfig(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		o       *operator
	}{
		{
			name:    "write config successfully",
			success: true,
			o:       mockOperator(emptyValidConfigPath, validConfig),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.o.writeConfig()
			assert.Equal(t, tc.success, err == nil)
			_ = os.Remove(tc.o.configFilePath)
		})
	}
}

func TestOperator_GetConfigItem(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		o           *operator
		key         string
		expectedVal any
	}{
		{
			name:        "get structured config item successfully type string",
			success:     true,
			o:           mockOperator(existValidConfigPath, validConfig),
			key:         "backends.pre.configs.host",
			expectedVal: "127.0.0.1",
		},
		{
			name:        "get structured config item successfully type int",
			success:     true,
			o:           mockOperator(existValidConfigPath, validConfig),
			key:         "backends.pre.configs.port",
			expectedVal: 3306,
		},
		{
			name:    "get structured config item successfully type pointer of struct",
			success: true,
			o:       mockOperator(existValidConfigPath, validConfig),
			key:     "backends.pre",
			expectedVal: &v1.BackendConfig{
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
			name:    "get structured config item successfully type map",
			success: true,
			o:       mockOperator(existValidConfigPath, validConfig),
			key:     "backends.prod.configs",
			expectedVal: map[string]any{
				v1.BackendGenericOssBucket: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.o.getConfigItem(tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedVal, val)
		})
	}
}

func TestOperator_GetEncodedConfigItem(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		o           *operator
		key         string
		expectedVal any
	}{
		{
			name:        "get encoding config item successfully type string",
			success:     true,
			o:           mockOperator(existValidConfigPath, validConfig),
			key:         "backends.pre.configs.host",
			expectedVal: "127.0.0.1",
		},
		{
			name:        "get encoding config item successfully type int",
			success:     true,
			o:           mockOperator(existValidConfigPath, validConfig),
			key:         "backends.pre.configs.port",
			expectedVal: "3306",
		},
		{
			name:        "get encoding config item successfully type map",
			success:     true,
			o:           mockOperator(existValidConfigPath, validConfig),
			key:         "backends.pre",
			expectedVal: `{"configs":{"dbName":"kusion","host":"127.0.0.1","port":3306,"user":"kk"},"type":"mysql"}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.o.getEncodedConfigItem(tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedVal, val)
		})
	}
}

func TestOperator_SetConfigItem(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		o              *operator
		key            string
		val            any
		expectedConfig *v1.Config
	}{
		{
			name:    "set config item successfully type string",
			success: true,
			o:       mockOperator(existValidConfigPath, nil),
			key:     "backends.dev.type",
			val:     v1.BackendTypeLocal,
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
		},
		{
			name:    "set config item successfully type int",
			success: true,
			o: mockOperator(existValidConfigPath, &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {Type: v1.BackendTypeMysql},
					},
				},
			}),
			key: "backends.pre.configs.port",
			val: 3306,
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlPort: 3306,
							},
						},
					},
				},
			},
		},
		{
			name:    "set config item successfully type struct",
			success: true,
			o:       mockOperator(existValidConfigPath, validConfig),
			key:     "backends.pre",
			val: &v1.BackendConfig{
				Type: v1.BackendTypeMysql,
				Configs: map[string]any{
					v1.BackendMysqlDBName: "kusion",
					v1.BackendMysqlUser:   "kk-tired",
					v1.BackendMysqlHost:   "127.0.0.1",
					v1.BackendMysqlPort:   3306,
				},
			},
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type: v1.BackendTypeLocal,
						},
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kusion",
								v1.BackendMysqlUser:   "kk-tired",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   3306,
							},
						},
						"prod": {
							Type: v1.BackendTypeS3,
							Configs: map[string]any{
								v1.BackendGenericOssBucket: "kusion",
							},
						},
					},
				},
			},
		},
		{
			name:    "set config item successfully type map",
			success: true,
			o:       mockOperator(existValidConfigPath, validConfig),
			key:     "backends.prod.configs",
			val: map[string]any{
				v1.BackendGenericOssBucket: "kk-so-tired",
			},
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type: v1.BackendTypeLocal,
						},
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kusion",
								v1.BackendMysqlUser:   "kk",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   3306,
							},
						},
						"prod": {
							Type: v1.BackendTypeS3,
							Configs: map[string]any{
								v1.BackendGenericOssBucket: "kk-so-tired",
							},
						},
					},
				},
			},
		},
		{
			name:           "failed to set config item invalid type",
			success:        false,
			o:              mockOperator(existValidConfigPath, nil),
			key:            "backends.dev.configs.path",
			val:            234,
			expectedConfig: nil,
		},
		{
			name:           "failed to set config item empty value",
			success:        false,
			o:              mockOperator(existValidConfigPath, nil),
			key:            "backends.dev.configs.path",
			val:            "",
			expectedConfig: nil,
		},
		{
			name:           "failed to set config item validate func failed",
			success:        false,
			o:              mockOperator(existValidConfigPath, nil),
			key:            "backends.dev.configs.path",
			val:            "/etc",
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.o.setConfigItem(tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				assert.Equal(t, tc.expectedConfig, tc.o.config)
			}
		})
	}
}

func TestOperator_setEncodedConfigItem(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		o              *operator
		key            string
		val            string
		expectedConfig *v1.Config
	}{
		{
			name:    "set config item successfully type string",
			success: true,
			o:       mockOperator(existValidConfigPath, nil),
			key:     "backends.dev.type",
			val:     v1.BackendTypeLocal,
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Type: v1.BackendTypeLocal},
					},
				},
			},
		},
		{
			name:    "set config item successfully type int",
			success: true,
			o: mockOperator(existValidConfigPath, &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {Type: v1.BackendTypeMysql},
					},
				},
			}),
			key: "backends.pre.configs.port",
			val: "3306",
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlPort: 3306,
							},
						},
					},
				},
			},
		},
		{
			name:    "set config item successfully type struct",
			success: true,
			o:       mockOperator(existValidConfigPath, validConfig),
			key:     "backends.pre",
			val:     `{"configs":{"dbName":"kusion","host":"127.0.0.1","port":3306,"user":"kk-tired"},"type":"mysql"}`,
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type: v1.BackendTypeLocal,
						},
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kusion",
								v1.BackendMysqlUser:   "kk-tired",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   3306,
							},
						},
						"prod": {
							Type: v1.BackendTypeS3,
							Configs: map[string]any{
								v1.BackendGenericOssBucket: "kusion",
							},
						},
					},
				},
			},
		},
		{
			name:    "set config item successfully type map",
			success: true,
			o:       mockOperator(existValidConfigPath, validConfig),
			key:     "backends.prod.configs",
			val:     `{"bucket":"kusion","region":"us-east-1"}`,
			expectedConfig: &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type: v1.BackendTypeLocal,
						},
						"pre": {
							Type: v1.BackendTypeMysql,
							Configs: map[string]any{
								v1.BackendMysqlDBName: "kusion",
								v1.BackendMysqlUser:   "kk",
								v1.BackendMysqlHost:   "127.0.0.1",
								v1.BackendMysqlPort:   3306,
							},
						},
						"prod": {
							Type: v1.BackendTypeS3,
							Configs: map[string]any{
								v1.BackendGenericOssBucket: "kusion",
								v1.BackendS3Region:         "us-east-1",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.o.setEncodedConfigItem(tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				assert.Equal(t, tc.expectedConfig, tc.o.config)
			}
		})
	}
}

func TestOperator_DeleteConfigItem(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		o              *operator
		key            string
		expectedConfig *v1.Config
	}{
		{
			name:    "delete config item successfully",
			success: true,
			o: mockOperator(existValidConfigPath, &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type:    v1.BackendTypeLocal,
							Configs: map[string]any{v1.BackendLocalPath: "/etc"},
						},
					},
				},
			}),
			key:            "backends.dev",
			expectedConfig: nil,
		},
		{
			name:    "failed to delete config item validateUnsetFunc failed",
			success: false,
			o: mockOperator(existValidConfigPath, &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type:    v1.BackendTypeLocal,
							Configs: map[string]any{v1.BackendLocalPath: "/etc"},
						},
					},
				},
			}),
			key:            "backends.dev",
			expectedConfig: nil,
		},
		{
			name:    "failed to delete config item unsupported key",
			success: false,
			o: mockOperator(existValidConfigPath, &v1.Config{
				Backends: &v1.BackendConfigs{
					Current: "dev",
					Backends: map[string]*v1.BackendConfig{
						"dev": {
							Type:    v1.BackendTypeLocal,
							Configs: map[string]any{v1.BackendLocalPath: "/etc"},
						},
					},
				},
			}),
			key:            "not_support",
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.o.deleteConfigItem(tc.key)
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				assert.Equal(t, tc.expectedConfig, tc.o.config)
			}
		})
	}
}

func TestGetConfigItemWithLaxType(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		config      *v1.Config
		key         string
		expectedVal any
	}{
		{
			name:        "failed to get config item with lax type empty config",
			success:     false,
			config:      nil,
			key:         "backends.current",
			expectedVal: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := getConfigItemWithLaxType(tc.config, tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedVal, val)
		})
	}
}

func TestTidyConfig(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		config         *v1.Config
		expectedConfig *v1.Config
	}{
		{
			name:    "tidy config successfully tidy config items",
			success: true,
			config: &v1.Config{
				Backends: &v1.BackendConfigs{
					Backends: map[string]*v1.BackendConfig{
						"dev": {Configs: map[string]any{}},
					},
				},
			},
			expectedConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tidyConfig(&tc.config)
			assert.Equal(t, tc.expectedConfig, tc.config)
		})
	}
}

func TestValidateConfigItem(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		info    *itemInfo
		key     string
		val     any
	}{
		{
			name:    "invalid config item empty value",
			success: false,
			config:  validConfig,
			info:    registeredItems["backends.current"],
			key:     "backends.current",
			val:     "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfigItem(tc.config, tc.info, tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestParseStructuredConfigItem(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		info    *itemInfo
		strVal  string
		val     any
	}{
		{
			name:    "parse structured config item successfully string",
			success: true,
			info:    registeredItems["backends.current"],
			strVal:  "dev",
			val:     "dev",
		},
		{
			name:    "parse structured config item successfully int",
			success: true,
			info:    registeredItems["backends.*.configs.port"],
			strVal:  "3306",
			val:     3306,
		},
		{
			name:    "parse structured config item successfully bool",
			success: true,
			info:    &itemInfo{false, nil, nil},
			strVal:  "true",
			val:     true,
		},
		{
			name:    "parse structured config item successfully struct ptr",
			success: true,
			info:    registeredItems["backends.*"],
			strVal:  `{"configs":{"dbName":"kusion","host":"127.0.0.1","port":3306,"user":"kk"},"type":"mysql"}`,
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
			name:    "parse structured config item successfully struct",
			success: true,
			info:    &itemInfo{v1.BackendConfig{}, nil, nil},
			strVal:  `{"configs":{"dbName":"kusion","host":"127.0.0.1","port":3306,"user":"kk"},"type":"mysql"}`,
			val: v1.BackendConfig{
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
			name:    "parse structured config item successfully map",
			success: true,
			info:    registeredItems["backends.*.configs"],
			strVal:  `{"bucket":"kusion"}`,
			val: map[string]any{
				v1.BackendGenericOssBucket: "kusion",
			},
		},
		{
			name:    "failed to parse structured config item int",
			success: false,
			info:    registeredItems["backends.*.configs.port"],
			strVal:  "not_valid_int",
			val:     nil,
		},
		{
			name:    "failed to parse structured config item bool",
			success: false,
			info:    &itemInfo{false, nil, nil},
			strVal:  "not_valid_bool",
			val:     nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := parseStructuredConfigItem(tc.info, tc.strVal)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.val, val)
		})
	}
}

func TestConvertToRegisteredKey(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		key           string
		registeredKey string
	}{
		{
			name:          "convert to registered key successfully keep the same",
			success:       true,
			key:           "backends.current",
			registeredKey: "backends.current",
		},
		{
			name:          "convert to registered key successfully convert backend name",
			success:       true,
			key:           "backends.dev.type",
			registeredKey: "backends.*.type",
		},
		{
			name:          "failed to convert to registered key unsupported key",
			success:       false,
			key:           "unsupported",
			registeredKey: "",
		},
		{
			name:          "failed to convert to registered key empty backend name",
			success:       false,
			key:           "backends..type",
			registeredKey: "",
		},
		{
			name:          "failed to convert to registered key invalid backend name",
			success:       false,
			key:           "backends.current.type",
			registeredKey: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			registeredKey, err := convertToRegisteredKey(registeredItems, tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.registeredKey, registeredKey)
		})
	}
}

func TestConvertCfgMap(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.Config
		cfg     map[string]any
	}{
		{
			name:    "convert config map successfully",
			success: true,
			config:  validConfig,
			cfg:     validCfg,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := convertToCfgMap(tc.config)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.cfg, cfg)
			config, err := convertFromCfgMap(tc.cfg)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.config, config)
		})
	}
}

func TestGetItemFromCfgMap(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		cfg         map[string]any
		key         string
		expectedVal any
	}{
		{
			name:        "get item from config map successfully type string",
			success:     true,
			cfg:         validCfg,
			key:         "backends.current",
			expectedVal: "dev",
		},
		{
			name:        "get item from config map successfully type int",
			success:     true,
			cfg:         validCfg,
			key:         "backends.pre.configs.port",
			expectedVal: 3306,
		},
		{
			name:    "get item from config map successfully type map",
			success: true,
			cfg:     validCfg,
			key:     "backends.prod",
			expectedVal: map[string]any{
				v1.BackendType: v1.BackendTypeS3,
				v1.BackendConfigItems: map[string]any{
					v1.BackendGenericOssBucket: "kusion",
				},
			},
		},
		{
			name:        "failed to get item from config map not exist value",
			success:     false,
			cfg:         validCfg,
			key:         "backends.dev.configs.path",
			expectedVal: nil,
		},
		{
			name:        "failed to get item from config map wrong key",
			success:     false,
			cfg:         validCfg,
			key:         "backends.stage.configs",
			expectedVal: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := getItemFromCfgMap(tc.cfg, tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedVal, val)
		})
	}
}

func TestSetItemFromCfgMap(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		cfg         map[string]any
		key         string
		val         any
		expectedCfg map[string]any
	}{
		{
			name:    "set item in config map successfully empty cfg",
			success: true,
			cfg:     map[string]any{},
			key:     "backends.current",
			val:     "dev",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
				},
			},
		},
		{
			name:    "set item in config map successfully exist cfg add new item",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"pre": map[string]any{
						"type": "mysql",
						"configs": map[string]any{
							"user": "kusion",
						},
					},
				},
			},
			key: "backends.pre.configs.dbName",
			val: "kusion",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"pre": map[string]any{
						"type": "mysql",
						"configs": map[string]any{
							"dbName": "kusion",
							"user":   "kusion",
						},
					},
				},
			},
		},
		{
			name:    "set item in config map successfully exist cfg add new item and tier",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"pre": map[string]any{
						"type": "mysql",
					},
				},
			},
			key: "backends.pre.configs.dbName",
			val: "kusion",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"pre": map[string]any{
						"type": "mysql",
						"configs": map[string]any{
							"dbName": "kusion",
						},
					},
				},
			},
		},
		{
			name:    "set item in config map successfully int",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"pre": map[string]any{
						"type": "mysql",
						"configs": map[string]any{
							"dbName": "kusion",
							"port":   8222,
						},
					},
				},
			},
			key: "backends.pre.configs.port",
			val: 3306,
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"pre": map[string]any{
						"type": "mysql",
						"configs": map[string]any{
							"port":   3306,
							"dbName": "kusion",
						},
					},
				},
			},
		},
		{
			name:    "set item in config map successfully struct",
			success: true,
			cfg:     map[string]any{},
			key:     "backends.pre",
			val: &v1.BackendConfig{
				Type: v1.BackendTypeMysql,
				Configs: map[string]any{
					v1.BackendMysqlDBName: "kusion",
					v1.BackendMysqlUser:   "kk",
					v1.BackendMysqlHost:   "127.0.0.1",
					v1.BackendMysqlPort:   3306,
				},
			},
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"pre": &v1.BackendConfig{
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
		{
			name:    "set item in config map successfully map",
			success: true,
			cfg:     map[string]any{},
			key:     "backends.prod.configs",
			val: map[string]any{
				"bucket": "kusion",
			},
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"prod": map[string]any{
						"configs": map[string]any{
							"bucket": "kusion",
						},
					},
				},
			},
		},
		{
			name:    "failed to set item in config map cannot assign type string",
			success: false,
			cfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
				},
			},
			key: "backends.current.test",
			val: "val",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
				},
			},
		},
		{
			name:    "failed to set item in config map cannot assign type slice",
			success: false,
			cfg: map[string]any{
				"test": []any{"val_1"},
			},
			key: "test.next",
			val: "val",
			expectedCfg: map[string]any{
				"test": []any{"val_1"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := setItemInCfgMap(tc.cfg, tc.key, tc.val)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedCfg, tc.cfg)
		})
	}
}

func TestDeleteItemInCfgMap(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		cfg         map[string]any
		key         string
		expectedCfg map[string]any
	}{
		{
			name:    "delete item in cfg map successfully end item",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
					"dev": map[string]any{
						"type": "local",
						"configs": map[string]any{
							"path": "etc",
						},
					},
				},
			},
			key: "backends.current",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"dev": map[string]any{
						"type": "local",
						"configs": map[string]any{
							"path": "etc",
						},
					},
				},
			},
		},
		{
			name:    "delete item in cfg map successfully end item in map",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
					"dev": map[string]any{
						"type": "local",
						"configs": map[string]any{
							"path": "etc",
						},
					},
				},
			},
			key: "backends.dev.configs.path",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
					"dev": map[string]any{
						"type":    "local",
						"configs": map[string]any{},
					},
				},
			},
		},
		{
			name:    "delete item in cfg map successfully middle item",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
					"dev": map[string]any{
						"type": "local",
						"configs": map[string]any{
							"path": "etc",
						},
					},
				},
			},
			key: "backends.dev",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
				},
			},
		},
		{
			name:    "delete item in cfg map successfully not exist key",
			success: true,
			cfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
					"dev": map[string]any{
						"type": "local",
						"configs": map[string]any{
							"path": "etc",
						},
					},
				},
			},
			key: "backends.current.notExist",
			expectedCfg: map[string]any{
				"backends": map[string]any{
					"current": "dev",
					"dev": map[string]any{
						"type": "local",
						"configs": map[string]any{
							"path": "etc",
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			deleteItemInCfgMap(tc.cfg, tc.key)
			assert.Equal(t, tc.expectedCfg, tc.cfg)
		})
	}
}
