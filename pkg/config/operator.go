package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/util/kfile"
)

const configFile = "config.yaml"

var (
	ErrEmptyConfigItem           = errors.New("empty config item")
	ErrConflictConfigItemType    = errors.New("type of the config item conflicts between saved and registered")
	ErrEmptyConfigItemKey        = errors.New("empty config item key")
	ErrEmptyConfigItemValue      = errors.New("empty config item value")
	ErrUnsupportedConfigItem     = errors.New("unsupported config item")
	ErrEmptyBackendName          = errors.New("backend name should not be empty")
	ErrInvalidBackendNameCurrent = errors.New("backend name should not be current")
)

// operator is used to execute the config management operation.
type operator struct {
	// configFilePath is the path of the kusion config file.
	configFilePath string

	// registeredItems includes the registered config items and corresponding registry info.
	registeredItems map[string]*itemInfo

	// config is the structured kusion config.
	config *v1.Config
}

// newOperator returns a kusion config operator.
func newOperator() (*operator, error) {
	kusionDataDir, err := kfile.KusionDataFolder()
	if err != nil {
		return nil, fmt.Errorf("get kusion data folder failed, %w", err)
	}

	o := &operator{
		configFilePath:  filepath.Join(kusionDataDir, configFile),
		registeredItems: newRegisteredItems(),
		config:          &v1.Config{},
	}

	if err = o.initDefaultConfig(); err != nil {
		return nil, err
	}
	return o, nil
}

// initDefaultConfig reads config from the config file and inits default config, which is called when new
// an operator. Now it inits the default backend.
func (o *operator) initDefaultConfig() error {
	if err := o.readConfig(); err != nil {
		return err
	}

	// set default backend config
	if o.config.Backends == nil {
		o.config.Backends = &v1.BackendConfigs{}
	}
	if o.config.Backends.Backends == nil {
		o.config.Backends.Backends = make(map[string]*v1.BackendConfig)
	}
	var needWrite bool
	defaultBackend := &v1.BackendConfig{Type: v1.BackendTypeLocal}
	if !reflect.DeepEqual(o.config.Backends.Backends[v1.DefaultBackendName], defaultBackend) {
		needWrite = true
		o.config.Backends.Backends[v1.DefaultBackendName] = defaultBackend
	}
	if o.config.Backends.Current == "" {
		needWrite = true
		o.config.Backends.Current = v1.DefaultBackendName
	}

	if needWrite {
		return o.writeConfig()
	}
	return nil
}

// readConfig reads config from config file.
func (o *operator) readConfig() error {
	content, err := os.ReadFile(o.configFilePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read kusion config file failed, %w", err)
	}

	cfg := &v1.Config{}
	if err = yaml.Unmarshal(content, cfg); err != nil {
		return fmt.Errorf("unmarshal kusion config failed, %w", err)
	}
	o.config = cfg
	return nil
}

// writeConfig writes config into config file.
func (o *operator) writeConfig() error {
	content, err := yaml.Marshal(o.config)
	if err != nil {
		return fmt.Errorf("marshal kusion config failed, %w", err)
	}
	if err = os.WriteFile(o.configFilePath, content, 0o640); err != nil {
		return fmt.Errorf("write kusion config file failed, %w", err)
	}
	return nil
}

// getConfigItem returns the value of the specified config item, which is structured and in the registered type.
func (o *operator) getConfigItem(key string) (any, error) {
	info, err := getRegisteredItemInfo(o.registeredItems, key)
	if err != nil {
		return nil, err
	}
	val, err := getConfigItemWithLaxType(o.config, key)
	if err != nil {
		return nil, err
	}

	switch val.(type) {
	case string, int, bool:
		if reflect.TypeOf(val) != reflect.TypeOf(info.zeroValue) {
			return nil, fmt.Errorf("%w, config item: %s", ErrConflictConfigItemType, key)
		}
		return val, nil
	default:
		var content []byte
		content, err = json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("encode value of config item %s failed, %w", key, err)
		}
		value := info.zeroValue
		// Cause json.Unmarshal will parse int to float, thus use yaml.Unmarshal, because
		// yaml is the superset of json.
		// ATTENTION! Please specify both the json key and yaml key for struct field explicitly
		// and keep the same, cause json use the field name but yaml lower case the initial
		// letter by default, which will cause error.
		if reflect.TypeOf(value).Kind() == reflect.Pointer {
			err = yaml.Unmarshal(content, value)
		} else {
			// ATTENTION! If using &value as the second input of yaml.Unmarshal, the type of it will
			// vanish, to map[string]any.
			valueAddr := reflect.New(reflect.TypeOf(value)).Interface()
			err = yaml.Unmarshal(content, valueAddr)
			value = reflect.ValueOf(valueAddr).Elem().Interface()
		}
		if err != nil {
			return nil, fmt.Errorf("decode value of config item %s failed, %w", key, err)
		}
		return value, nil
	}
}

// getEncodedConfigItem returns the value of the specified config item, which is encoded and in the type
// of string.
func (o *operator) getEncodedConfigItem(key string) (string, error) {
	_, err := getRegisteredItemInfo(o.registeredItems, key)
	if err != nil {
		return "", err
	}
	val, err := getConfigItemWithLaxType(o.config, key)
	if err != nil {
		return "", err
	}

	// encode val to string, which is the inverse process of decoding
	switch value := val.(type) {
	case string:
		return value, nil
	case int:
		return fmt.Sprintf("%d", value), nil
	case bool:
		return fmt.Sprintf("%t", value), nil
	default:
		var content []byte
		content, err = json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("encode value of config %s failed, %w", key, err)
		}
		return string(content), nil
	}
}

// setConfigItem sets the value of the config item specified by the key, where the value is structured and must
// in the registered type.
func (o *operator) setConfigItem(key string, value any) error {
	info, err := getRegisteredItemInfo(o.registeredItems, key)
	if err != nil {
		return err
	}
	var config *v1.Config
	if config, err = setItemInConfig(o.config, info, key, value); err != nil {
		return err
	} else {
		o.config = config
	}
	return nil
}

// setEncodedConfigItem sets the value of the config item specified by the key, where the value is encoded and
// in the type of string.
func (o *operator) setEncodedConfigItem(key string, strValue string) error {
	info, err := getRegisteredItemInfo(o.registeredItems, key)
	if err != nil {
		return err
	}
	value, err := parseStructuredConfigItem(info, strValue)
	if err != nil {
		return fmt.Errorf("decode valu of %s faield, %w", key, err)
	}
	var config *v1.Config
	if config, err = setItemInConfig(o.config, info, key, value); err != nil {
		return err
	} else {
		o.config = config
	}
	return nil
}

// deleteConfigItem deletes the config item specified by the key.
func (o *operator) deleteConfigItem(key string) error {
	info, err := getRegisteredItemInfo(o.registeredItems, key)
	if err != nil {
		return err
	}
	if info.validateUnsetFunc != nil {
		if err = info.validateUnsetFunc(o.config, key); err != nil {
			return err
		}
	}

	cfg, err := convertToCfgMap(o.config)
	if err != nil {
		return err
	}
	deleteItemInCfgMap(cfg, key)
	config, err := convertFromCfgMap(cfg)
	if err != nil {
		return err
	}

	// clean unexpected empty data block
	tidyConfig(&config)
	o.config = config
	return nil
}

// getConfigItemWithLaxType gets the value of the specified config item from config, where the type of the
// value is that in the converted config map.
func getConfigItemWithLaxType(config *v1.Config, key string) (any, error) {
	cfg, err := convertToCfgMap(config)
	if err != nil {
		return nil, err
	}
	return getItemFromCfgMap(cfg, key)
}

// setItemInConfig updates the value of the config item specified by the key in config. If succeeded, return
// the updated config.
// ATTENTION! Cause the input config could be nil, change it or its value may both could result in assignment failure.
func setItemInConfig(config *v1.Config, info *itemInfo, key string, value any) (*v1.Config, error) {
	if err := validateConfigItem(config, info, key, value); err != nil {
		return nil, err
	}

	cfg, err := convertToCfgMap(config)
	if err != nil {
		return nil, err
	}
	if err = setItemInCfgMap(cfg, key, value); err != nil {
		return nil, err
	}
	return convertFromCfgMap(cfg)
}

// tidyConfig is used to clean dirty empty block.
func tidyConfig(configAddr **v1.Config) {
	config := *configAddr

	if config.Backends != nil {
		for name, cfg := range config.Backends.Backends {
			if cfg != nil && len(cfg.Configs) == 0 {
				cfg.Configs = nil
			}
			if cfg == nil || reflect.ValueOf(*cfg).IsZero() {
				delete(config.Backends.Backends, name)
			}
		}
		if len(config.Backends.Backends) == 0 {
			config.Backends.Backends = make(map[string]*v1.BackendConfig)
		}
	}

	*configAddr = config
}

// validateConfigItem checks the config item value is valid or not.
func validateConfigItem(config *v1.Config, info *itemInfo, key string, value any) error {
	if reflect.ValueOf(value).IsZero() {
		return ErrEmptyConfigItemValue
	}
	if info.ValidateSetFunc != nil {
		return info.ValidateSetFunc(config, key, value)
	}
	return nil
}

// parseStructuredConfigItem decodes the config item value from string to the registered type and check
// its validation.
func parseStructuredConfigItem(info *itemInfo, strValue string) (any, error) {
	if len(strValue) == 0 {
		return nil, ErrEmptyConfigItemValue
	}

	value := info.zeroValue
	var err error
	switch value.(type) {
	case string:
		value = strValue
	case int:
		value, err = strconv.Atoi(strValue)
		if err != nil {
			return nil, ErrNotInt
		}
	case bool:
		value, err = strconv.ParseBool(strValue)
		if err != nil {
			return nil, ErrNotBool
		}
	default:
		if reflect.TypeOf(value).Kind() == reflect.Pointer {
			err = yaml.Unmarshal([]byte(strValue), value)
		} else {
			valueAddr := reflect.New(reflect.TypeOf(value)).Interface()
			err = yaml.Unmarshal([]byte(strValue), valueAddr)
			value = reflect.ValueOf(valueAddr).Elem().Interface()
		}
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// getRegisteredItemInfo returns the registered info of the config key. If the config key is not registered,
// return error.
func getRegisteredItemInfo(registeredItems map[string]*itemInfo, key string) (*itemInfo, error) {
	if key == "" {
		return nil, ErrEmptyConfigItemKey
	}
	registeredKey, err := convertToRegisteredKey(registeredItems, key)
	if err != nil {
		return nil, err
	}
	return registeredItems[registeredKey], nil
}

func convertToRegisteredKey(registeredItems map[string]*itemInfo, key string) (string, error) {
	fields := strings.Split(key, ".")

	var registeredKey string
	var err error
	switch fields[0] {
	case v1.ConfigBackends:
		if registeredKey, err = convertBackendKey(key); err != nil {
			return "", err
		}
	default:
		return "", ErrUnsupportedConfigItem
	}

	if _, ok := registeredItems[registeredKey]; !ok {
		return "", ErrUnsupportedConfigItem
	}
	return registeredKey, nil
}

func convertBackendKey(key string) (string, error) {
	fields := strings.Split(key, ".")
	if len(fields) < 2 || len(fields) > 4 {
		return "", fmt.Errorf("%w, %s", ErrUnsupportedConfigItem, key)
	}
	if fields[1] == v1.BackendCurrent && len(fields) == 2 {
		return key, nil
	}
	if fields[1] == v1.BackendCurrent {
		return "", ErrInvalidBackendNameCurrent
	}
	if fields[1] == "" {
		return "", ErrEmptyBackendName
	}
	fields[1] = "*"
	registeredKey := strings.Join(fields, ".")
	return registeredKey, nil
}

func convertToCfgMap(config *v1.Config) (cfg map[string]any, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("convert config to map failed, %w", err)
		}
	}()

	content, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	cfg = make(map[string]any)
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func convertFromCfgMap(cfg map[string]any) (config *v1.Config, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("convert config from map failed, %w", err)
		}
	}()

	content, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	config = &v1.Config{}
	if err = yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func getItemFromCfgMap(cfg map[string]any, key string) (any, error) {
	fields := strings.Split(key, ".")
	var obj, value any
	obj = cfg
	for i, field := range fields {
		switch object := obj.(type) {
		case map[string]any:
			val, ok := object[field]
			if ok && i != len(fields)-1 {
				obj = val
			} else if ok && i == len(fields)-1 {
				value = val
			} else {
				return nil, ErrEmptyConfigItem
			}
		default:
			return nil, ErrEmptyConfigItem
		}
	}
	return value, nil
}

func setItemInCfgMap(cfg map[string]any, key string, value any) error {
	fields := strings.Split(key, ".")
	var obj any
	obj = cfg
	for i, field := range fields {
		switch object := obj.(type) {
		case map[string]any:
			val, ok := object[field]
			if i == len(fields)-1 {
				object[field] = value
			} else if ok {
				obj = val
			} else {
				object[field] = make(map[string]any)
				obj = object[field]
			}
		default:
			return fmt.Errorf("config item %s is not assignable, with invalid type", strings.Join(fields[:i+1], "."))
		}
	}
	return nil
}

func deleteItemInCfgMap(cfg map[string]any, key string) {
	fields := strings.Split(key, ".")
	var obj any
	obj = cfg
	for i, field := range fields {
		switch object := obj.(type) {
		case map[string]any:
			val, ok := object[field]
			if ok && i != len(fields)-1 {
				obj = val
			} else if ok && i == len(fields)-1 {
				delete(object, field)
			} else {
				break
			}
		default:
			break
		}
	}
}
