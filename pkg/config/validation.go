package config

import (
	"errors"
	"fmt"
	"strings"

	v1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
)

type (
	validateFunc       func(config *v1.Config, key string, val any) error
	validateDeleteFunc func(config *v1.Config, key string) error
)

var (
	ErrNotExistCurrentBackend     = errors.New("cannot assign current to not exist backend")
	ErrUnsetDefaultCurrentBackend = errors.New("cannot unset default current backend")
	ErrInUseCurrentBackend        = errors.New("unset in-use current backend")
	ErrUnsupportedBackendType     = errors.New("unsupported backend type")
	ErrNonEmptyBackendConfigItems = errors.New("non-empty backend config items")
	ErrEmptyBackendType           = errors.New("empty backend type")
	ErrConflictBackendType        = errors.New("conflict backend type")
	ErrInvalidBackNameDefault     = errors.New("backend name should not be default")
)

// validateSetCurrentBackend is used to check that setting the current backend is valid or not.
func validateSetCurrentBackend(config *v1.Config, _ string, val any) error {
	current, _ := val.(string)
	_, ok := config.Backends.Backends[current]
	if ok {
		return nil
	}
	return ErrNotExistCurrentBackend
}

// validateUnsetCurrentBackend is used to check that unsetting the current backend is valid or not.
func validateUnsetCurrentBackend(config *v1.Config, _ string) error {
	if config.Backends.Current == v1.DefaultBackendName {
		return ErrUnsetDefaultCurrentBackend
	}
	return nil
}

// validateSetBackendConfig is used to check that setting the backend config is valid or not.
func validateSetBackendConfig(_ *v1.Config, key string, val any) error {
	if err := checkNotDefaultBackendName(parseBackendName(key)); err != nil {
		return err
	}
	config, _ := val.(*v1.BackendConfig)
	return checkBackendConfig(config)
}

// validateUnsetBackendConfig is used to check that unsetting the backend config is valid or not.
func validateUnsetBackendConfig(config *v1.Config, key string) error {
	backendName := parseBackendName(key)
	if err := checkNotDefaultBackendName(backendName); err != nil {
		return err
	}
	if config.Backends.Current == backendName {
		return fmt.Errorf("%w, cannot unset config of backend %s cause it's current backend", ErrInUseCurrentBackend, config.Backends.Current)
	}
	return nil
}

// validateSetBackendType is used to check that setting the backend type is valid or not.
func validateSetBackendType(config *v1.Config, key string, val any) error {
	backendType, _ := val.(string)
	if backendType != v1.BackendTypeLocal && backendType != v1.BackendTypeOss && backendType != v1.BackendTypeS3 {
		return ErrUnsupportedBackendType
	}

	backendName := parseBackendName(key)
	if err := checkNotDefaultBackendName(backendName); err != nil {
		return err
	}
	if config.Backends.Backends[backendName] == nil {
		return nil
	}
	if config.Backends.Backends[backendName].Type != backendType && len(config.Backends.Backends[backendName].Configs) != 0 {
		return fmt.Errorf("%w, cannot assign backend %s from type %s to %s with non-empyt config", ErrConflictBackendType, backendName, config.Backends.Backends[backendName].Type, backendType)
	}
	return nil
}

// validateUnsetBackendType is used to check that unsetting the backend config is valid or not.
func validateUnsetBackendType(config *v1.Config, key string) error {
	backendName := parseBackendName(key)
	if err := checkNotDefaultBackendName(backendName); err != nil {
		return err
	}
	if config.Backends.Backends[backendName] == nil {
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

// validateSetBackendConfigItems is used to check that setting the backend config items is valid or not.
func validateSetBackendConfigItems(config *v1.Config, key string, val any) error {
	configItems, _ := val.(map[string]any)
	backendName := parseBackendName(key)
	if err := checkNotDefaultBackendName(backendName); err != nil {
		return err
	}
	if config.Backends.Backends[backendName] == nil || config.Backends.Backends[backendName].Type == "" {
		return ErrEmptyBackendType
	}
	bkConfig := &v1.BackendConfig{
		Type:    config.Backends.Backends[backendName].Type,
		Configs: configItems,
	}
	return checkBackendConfig(bkConfig)
}

// validateUnsetBackendConfigItems is used to check that unsetting the backend config items is valid or not.
func validateUnsetBackendConfigItems(_ *v1.Config, key string) error {
	return checkNotDefaultBackendName(parseBackendName(key))
}

// validateSetLocalBackendItem is used to check that setting the config item for local-type backend is valid or not.
func validateSetLocalBackendItem(config *v1.Config, key string, _ any) error {
	if err := checkNotDefaultBackendName(parseBackendName(key)); err != nil {
		return err
	}
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeLocal)
}

// validateUnsetLocalBackendItem is used to check that unsetting the config item for local-type backend is valid or not.
func validateUnsetLocalBackendItem(_ *v1.Config, key string) error {
	return checkNotDefaultBackendName(parseBackendName(key))
}

// validateSetGenericOssBackendItem is used to check that setting the generic config of oss/s3-type backend is valid or not.
func validateSetGenericOssBackendItem(config *v1.Config, key string, _ any) error {
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeOss, v1.BackendTypeS3)
}

// validateSetS3BackendItem is used to check that setting the bucket of s3-type backend is valid or not.
func validateSetS3BackendItem(config *v1.Config, key string, _ any) error {
	return checkBackendTypeForBackendItem(config, key, v1.BackendTypeS3)
}

// checkBackendConfig is used to check that setting the backend config is valid or not, which is called
// validateSetBackendConfig and validateSetBackendConfigItems.
func checkBackendConfig(config *v1.BackendConfig) error {
	if err := checkBasalBackendConfig(config); err != nil {
		return err
	}

	switch config.Type {
	case v1.BackendTypeOss:
		ossBackend := config.ToOssBackend()
		if err := storages.ValidateOssConfigFromFile(ossBackend); err != nil {
			return err
		}
	case v1.BackendTypeS3:
		s3Backend := config.ToS3Backend()
		if err := storages.ValidateS3ConfigFromFile(s3Backend); err != nil {
			return err
		}
	}
	return nil
}

// checkBasalBackendConfig does basal validation of the backend config. Besides used when setting backend
// config, it's also used to check the validation of the current backend config
func checkBasalBackendConfig(config *v1.BackendConfig) error {
	switch config.Type {
	case v1.BackendTypeLocal:
		items := map[string]checkTypeFunc{
			v1.BackendLocalPath: checkString,
		}
		if err := checkBasalBackendConfigItems(config, items); err != nil {
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
		if err := checkBasalBackendConfigItems(config, items); err != nil {
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
		if err := checkBasalBackendConfigItems(config, items); err != nil {
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
	if config.Backends.Backends[backendName] == nil || config.Backends.Backends[backendName].Type == "" {
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

// checkNotDefaultBackendName returns error if the backend name is default.
func checkNotDefaultBackendName(name string) error {
	if name == v1.DefaultBackendName {
		return ErrInvalidBackNameDefault
	}
	return nil
}

// parseBackendName parses the backend name from the config key, the key is like "backends.dev.configs.bucket",
// "backends.dev.type", "backends.dev"
func parseBackendName(key string) string {
	fields := strings.Split(key, ".")
	if len(fields) < 2 {
		return ""
	}
	return fields[1]
}

// parseBackendItem parses the backend config item from the config key, the key is like "backends.dev.configs.bucket"
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
