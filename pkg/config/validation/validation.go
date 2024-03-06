package validation

import (
	"errors"
	"fmt"
	"strings"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

type (
	ValidateFunc      func(config *v1.Config, key string, val any) error
	ValidateUnsetFunc func(config *v1.Config, key string) error
)

var (
	ErrUnexpectedInvalidConfig    = errors.New("unexpected invalid config")
	ErrNotExistCurrentBackend     = errors.New("cannot assign current to not exist backend")
	ErrInUseCurrentBackend        = errors.New("unset in-use current backend")
	ErrUnsupportedBackendType     = errors.New("unsupported backend type")
	ErrNonEmptyBackendConfigItems = errors.New("non-empty backend config items")
	ErrEmptyBackendType           = errors.New("empty backend type")
	ErrConflictBackendType        = errors.New("conflict backend type")
	ErrInvalidBackendMysqlPort    = errors.New("backend mysql port must be between 1 and 65535")
)

// ValidateConfig is used to check the config is valid or not, where the invalidation comes from the unexpected
// manual modification.
func ValidateConfig(config *v1.Config) error {
	if config == nil {
		return nil
	}

	// validate backends configuration
	backends := config.Backends
	if backends == nil {
		return nil
	}
	if backends.Current != "" && len(backends.Backends) == 0 {
		return fmt.Errorf("%w, non-empty current backend name %s but empty backends", ErrUnexpectedInvalidConfig, backends.Current)
	}
	for name, backendConfig := range backends.Backends {
		if name == "" || name == "current" {
			return fmt.Errorf("%w, invalid backend name %s", ErrUnexpectedInvalidConfig, name)
		}
		if backendConfig == nil {
			return fmt.Errorf("%w, empty backend config with name %s", ErrUnexpectedInvalidConfig, name)
		}
		if backendConfig.Type == "" && len(backendConfig.Configs) != 0 {
			return fmt.Errorf("%w, empty backend config item but non-empty type with name %s", ErrUnexpectedInvalidConfig, name)
		}
		if err := validateBasalBackendConfig(backendConfig); err != nil {
			return fmt.Errorf("%w, %v", ErrUnexpectedInvalidConfig, err)
		}
	}
	return nil
}

// ValidateCurrentBackend is used to check that setting the current backend is valid or not.
func ValidateCurrentBackend(config *v1.Config, _ string, val any) error {
	current, _ := val.(string)
	if config != nil && config.Backends != nil && config.Backends.Backends != nil {
		_, ok := config.Backends.Backends[current]
		if ok {
			return nil
		}
	}
	return ErrNotExistCurrentBackend
}

// ValidateBackendConfig is used to check that setting the backend config is valid or not.
func ValidateBackendConfig(_ *v1.Config, _ string, val any) error {
	backendConfig, _ := val.(*v1.BackendConfig)
	return validateBackendConfig(backendConfig)
}

// ValidateUnsetBackendConfig is used to check that unsetting the backend config is valid or not.
func ValidateUnsetBackendConfig(config *v1.Config, key string) error {
	if config == nil || config.Backends == nil {
		return nil
	}
	if config.Backends.Current == parseBackendName(key) {
		return fmt.Errorf("%w, cannot unset config of backend %s cause it's current backend", ErrInUseCurrentBackend, config.Backends.Current)
	}
	return nil
}

// ValidateBackendType is used to check that setting the backend type is valid or not.
func ValidateBackendType(config *v1.Config, key string, val any) error {
	backendType, _ := val.(string)
	if backendType != v1.BackendTypeLocal && backendType != v1.BackendTypeMysql && backendType != v1.BackendTypeOss && backendType != v1.BackendTypeS3 {
		return ErrUnsupportedBackendType
	}

	backendName := parseBackendName(key)
	if config == nil || config.Backends == nil || config.Backends.Backends == nil || config.Backends.Backends[backendName] == nil {
		return nil
	}
	if config.Backends.Backends[backendName].Type != backendType && len(config.Backends.Backends[backendName].Configs) != 0 {
		return fmt.Errorf("%w, cannot assign backend %s from type %s to %s with non-empyt config", ErrConflictBackendType, backendName, config.Backends.Backends[backendName].Type, backendType)
	}
	return nil
}

// ValidateUnsetBackendType is used to check that unsetting the backend config is valid or not.
func ValidateUnsetBackendType(config *v1.Config, key string) error {
	backendName := parseBackendName(key)
	if config == nil || config.Backends == nil || config.Backends.Backends == nil || config.Backends.Backends[backendName] == nil {
		return nil
	}
	if len(config.Backends.Backends[backendName].Configs) != 0 {
		return fmt.Errorf("%w, cannot unset type backend of %s whose config items is not empty", ErrNonEmptyBackendConfigItems, backendName)
	}
	if config.Backends.Current == backendName {
		return fmt.Errorf("%w, cannot unset config of backend %s cause it's current backend", ErrInUseCurrentBackend, config.Backends.Current)
	}
	return nil
}

// ValidateBackendConfigItems is used to check that setting the backend config items is valid or not.
func ValidateBackendConfigItems(config *v1.Config, key string, val any) error {
	configItems, _ := val.(map[string]any)
	backendName := parseBackendName(key)
	if config == nil || config.Backends == nil || config.Backends.Backends == nil || config.Backends.Backends[backendName] == nil || config.Backends.Backends[backendName].Type == "" {
		return ErrEmptyBackendType
	}
	backendConfig := &v1.BackendConfig{
		Type:    config.Backends.Backends[backendName].Type,
		Configs: configItems,
	}
	return validateBackendConfig(backendConfig)
}

// ValidateLocalBackendItem is used to check that setting the config item for local-type backend is valid or not.
func ValidateLocalBackendItem(config *v1.Config, key string, _ any) error {
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeLocal)
}

// ValidateMysqlBackendItem is used to check that setting the config item for mysql-type backend is valid or not.
func ValidateMysqlBackendItem(config *v1.Config, key string, _ any) error {
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeMysql)
}

// ValidateMysqlBackendPort is used to check that setting the port of mysql-type backend is valid or not.
func ValidateMysqlBackendPort(config *v1.Config, key string, val any) error {
	if err := ValidateMysqlBackendItem(config, key, val); err != nil {
		return err
	}
	port, _ := val.(int)
	if port < 1 || port > 65535 {
		return ErrInvalidBackendMysqlPort
	}
	return nil
}

// ValidateGenericOssBackendItem is used to check that setting the generic config of oss/s3-type backend is valid or not.
func ValidateGenericOssBackendItem(config *v1.Config, key string, _ any) error {
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeOss, v1.BackendTypeS3)
}

// ValidateS3BackendItem is used to check that setting the bucket of s3-type backend is valid or not.
func ValidateS3BackendItem(config *v1.Config, key string, _ any) error {
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeS3)
}

// validateBackendConfig is used to check that setting the backend config is valid or not, which is called
// ValidateBackendConfig and ValidateBackendConfigItems.
func validateBackendConfig(backendConfig *v1.BackendConfig) error {
	if err := validateBasalBackendConfig(backendConfig); err != nil {
		return err
	}

	switch backendConfig.Type {
	case v1.BackendTypeMysql:
		mysqlBackend := backendConfig.ToMysqlBackend()
		if err := ValidateMysqlConfig(mysqlBackend); err != nil {
			return err
		}
	case v1.BackendTypeOss:
		ossBackend := backendConfig.ToOssBackend()
		if err := ValidateGenericObjectStorageConfig(ossBackend.GenericBackendObjectStorageConfig); err != nil {
			return err
		}
	case v1.BackendTypeS3:
		ossBackend := backendConfig.ToS3Backend()
		if err := ValidateGenericObjectStorageConfig(ossBackend.GenericBackendObjectStorageConfig); err != nil {
			return err
		}
	}
	return nil
}

// validateBasalBackendConfig does basal validation of the backend config. Besides used when setting backend
// config, it's also used to check the validation of the current backend config
func validateBasalBackendConfig(backendConfig *v1.BackendConfig) error {
	switch backendConfig.Type {
	case v1.BackendTypeLocal:
		items := map[string]checkTypeFunc{
			v1.BackendLocalPath: checkString,
		}
		if err := checkBasalBackendConfigItems(backendConfig, items); err != nil {
			return err
		}
	case v1.BackendTypeMysql:
		items := map[string]checkTypeFunc{
			v1.BackendMysqlDBName:   checkString,
			v1.BackendMysqlUser:     checkString,
			v1.BackendMysqlPassword: checkString,
			v1.BackendMysqlHost:     checkString,
			v1.BackendMysqlPort:     checkInt,
		}
		if err := checkBasalBackendConfigItems(backendConfig, items); err != nil {
			return err
		}
	case v1.BackendTypeOss:
		items := map[string]checkTypeFunc{
			v1.BackendGenericOssEndpoint: checkString,
			v1.BackendGenericOssAK:       checkString,
			v1.BackendGenericOssSK:       checkString,
			v1.BackendGenericOssBucket:   checkString,
			v1.BackendGenericOssPrefix:   checkString,
		}
		if err := checkBasalBackendConfigItems(backendConfig, items); err != nil {
			return err
		}
	case v1.BackendTypeS3:
		items := map[string]checkTypeFunc{
			v1.BackendGenericOssEndpoint: checkString,
			v1.BackendGenericOssAK:       checkString,
			v1.BackendGenericOssSK:       checkString,
			v1.BackendGenericOssBucket:   checkString,
			v1.BackendGenericOssPrefix:   checkString,
			v1.BackendS3Region:           checkString,
		}
		if err := checkBasalBackendConfigItems(backendConfig, items); err != nil {
			return err
		}
	default:
		return ErrUnsupportedBackendType
	}
	return nil
}

// checkBasalBackendConfigItems is used to check type of the backend config and whether it's the supported item.
func checkBasalBackendConfigItems(backend *v1.BackendConfig, items map[string]checkTypeFunc) error {
	for configItem, configValue := range backend.Configs {
		checkType, ok := items[configItem]
		if !ok {
			return fmt.Errorf("do not support %s for backend with type %s", configItem, backend.Type)
		}
		if err := checkType(configValue); err != nil {
			return fmt.Errorf("value of %s with backend type %s is %w", configItem, backend.Type, err)
		}
	}
	return nil
}

// checkBackendTypeForBackendItem checks the backend type when setting backend config item
func checkBackendTypeForBackendItem(config *v1.Config, key string, backendTypes ...string) error {
	backendName := parseBackendName(key)
	if config == nil || config.Backends == nil || config.Backends.Backends == nil || config.Backends.Backends[backendName] == nil || config.Backends.Backends[backendName].Type == "" {
		return ErrEmptyBackendType
	}

	validType := false
	for _, backendType := range backendTypes {
		if config.Backends.Backends[backendName].Type == backendType {
			validType = true
			break
		}
	}
	if !validType {
		itemName := parseBackendItem(key)
		return fmt.Errorf("%w, %s cannot assign to backend %s with type %s", ErrConflictBackendType, itemName, backendName, config.Backends.Backends[backendName].Type)
	}
	return nil
}

// parseBackendName parses the backend name from the config key, the key is like "backends.dev.configs.dbName",
// "backends.dev.type", "backends.dev"
func parseBackendName(key string) string {
	fields := strings.Split(key, ".")
	if len(fields) < 2 {
		return ""
	}
	return fields[1]
}

// parseBackendItem parses the backend config item from the config key, the key is like "backends.dev.configs.dbName"
func parseBackendItem(key string) string {
	fields := strings.Split(key, ".")
	if len(fields) != 4 {
		return ""
	}
	return fields[3]
}

type checkTypeFunc func(val any) error

var (
	ErrNotBool   = errors.New("not bool type")
	ErrNotInt    = errors.New("not int type")
	ErrNotString = errors.New("not string type")
)

func checkString(val any) error {
	if _, ok := val.(string); !ok {
		return ErrNotString
	}
	return nil
}

func checkInt(val any) error {
	if _, ok := val.(int); !ok {
		return ErrNotInt
	}
	return nil
}

// todo: the following funcs should be moved to package backend later

var (
	ErrEmptyMysqlDBName = errors.New("empty db name")
	ErrEmptyMysqlUser   = errors.New("empty mysql db user")
	ErrEmptyMysqlHost   = errors.New("empty mysql host")
	ErrInvalidMysqlPort = errors.New("mysql port must be between 1 and 65535")
	ErrEmptyBucket      = errors.New("empty bucket")
)

// ValidateMysqlConfig is used to validate v1.BackendMysqlConfig is valid or not.
func ValidateMysqlConfig(config *v1.BackendMysqlConfig) error {
	if config.DBName == "" {
		return ErrEmptyMysqlDBName
	}
	if config.User == "" {
		return ErrEmptyMysqlUser
	}
	if config.Host == "" {
		return ErrEmptyMysqlHost
	}
	if config.Port != 0 && (config.Port < 1 || config.Port > 65535) {
		return ErrInvalidMysqlPort
	}
	return nil
}

// ValidateGenericObjectStorageConfig is used to validate v1.BackendOssConfig and v1.BackendS3Config is
// valid or not, where the sensitive data items set as environment variables are not included.
func ValidateGenericObjectStorageConfig(config *v1.GenericBackendObjectStorageConfig) error {
	if config.Bucket == "" {
		return ErrEmptyBucket
	}
	return nil
}
