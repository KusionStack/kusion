package config

import (
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

const (
	backendCurrent            = v1.ConfigBackends + "." + v1.BackendCurrent
	backendConfig             = v1.ConfigBackends + "." + "*"
	backendConfigType         = backendConfig + "." + v1.BackendType
	backendConfigItems        = backendConfig + "." + v1.BackendConfigItems
	backendLocalPath          = backendConfigItems + "." + v1.BackendLocalPath
	backendGenericOssEndpoint = backendConfigItems + "." + v1.BackendGenericOssEndpoint
	backendGenericOssAK       = backendConfigItems + "." + v1.BackendGenericOssAK
	backendGenericOssSK       = backendConfigItems + "." + v1.BackendGenericOssSK
	backendGenericOssBucket   = backendConfigItems + "." + v1.BackendGenericOssBucket
	backendGenericOssPrefix   = backendConfigItems + "." + v1.BackendGenericOssPrefix
	backendS3Region           = backendConfigItems + "." + v1.BackendS3Region
)

func newRegisteredItems() map[string]*itemInfo {
	return map[string]*itemInfo{
		backendCurrent:            {"", validateSetCurrentBackend, validateUnsetCurrentBackend},
		backendConfig:             {&v1.BackendConfig{}, validateSetBackendConfig, validateUnsetBackendConfig},
		backendConfigType:         {"", validateSetBackendType, validateUnsetBackendType},
		backendConfigItems:        {map[string]any{}, validateSetBackendConfigItems, validateUnsetBackendConfigItems},
		backendLocalPath:          {"", validateSetLocalBackendItem, validateUnsetLocalBackendItem},
		backendGenericOssEndpoint: {"", validateSetGenericOssBackendItem, nil},
		backendGenericOssAK:       {"", validateSetGenericOssBackendItem, nil},
		backendGenericOssSK:       {"", validateSetGenericOssBackendItem, nil},
		backendGenericOssBucket:   {"", validateSetGenericOssBackendItem, nil},
		backendGenericOssPrefix:   {"", validateSetGenericOssBackendItem, nil},
		backendS3Region:           {"", validateSetS3BackendItem, nil},
	}
}

// itemInfo includes necessary information of the config item, which is used when getting, setting and unsetting
// the config item.
type itemInfo struct {
	// zeroValue is the zero value of the type that the config item will be parsed from string to. Support string,
	// int, bool, map, slice, struct, and pointer of struct, the parser rule is shown as below:
	//	- string: keep the same.
	//	- int: calling strconv.Atoi to decode, and fmt.Sprintf to encode, e.g. "45" is valid, parsed to 45,
	// 	- bool: calling strconv.ParseBool to decode, and fmt.Sprintf to encode, e.g. "true" is valid, parsed to true.
	//	- slice, map, struct(ptr of struct): calling yaml.Unmarshal to decode, and json.Marshal to encode, e.g.
	//		map[string]any{}. The zeroValue must be initialized, nil is invalid. For slice, map and struct, the
	//		address of the zeroValue will be used as the input of yaml.Unmarshal; for ptr of struct, use the
	//		zeroValue itself.
	// ATTENTION! For other unsupported types, calling json.Unmarshal to do the parse job, unexpected error or panic
	// may happen. Please do not use them.
	zeroValue any

	// ValidateSetFunc is used to check the config item is valid or not to set, calling before executing real
	// config setting. The unregistered config item, empty item value and invalid item value type is forbidden
	// by config operator by default, which are unnecessary to check in the ValidateSetFunc.
	// Please do not do any real setting job in the ValidateSetFunc.
	ValidateSetFunc validateFunc

	// validateUnsetFunc is used to check the config item is valid or not to unset, calling before executing
	// real config unsetting.
	// Please do not do any real unsetting job in the validateUnsetFunc.
	validateUnsetFunc validateDeleteFunc
}
