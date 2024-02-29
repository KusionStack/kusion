package backend

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	backendinit "kusionstack.io/kusion/pkg/engine/backend/init"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	ErrEmptyBackendType             = errors.New("empty --backend-type")
	ErrUnsupportedBackendType       = errors.New("unsupported --backend-type")
	ErrInvalidBackendConfigFormat   = errors.New("invalid --backend-config format, should with format [\"<configKey>=<configValue>\"]")
	ErrEmptyBackendConfigKey        = errors.New("empty config key in --backend-config item")
	ErrEmptyBackendConfigValue      = errors.New("empty config value in --backend-config item")
	ErrUnsupportedBackendConfigItem = errors.New("unsupported --backend-config item")
	ErrNotSupportBackendConfig      = errors.New("do not support --backend-config")
)

// BackendOptions is the kusion cli backend override config
type BackendOptions struct {
	// Type is the type of backend, currently supported:
	//		local	- state is stored to a local file
	//		mysql	- state is stored to mysql
	//    	oss 	- state is stored to aliyun oss
	//		s3 		- state is stored to aws s3
	//		http  	- state is stored by http service
	Type string

	// Config is a group of configurations of the specified type backend, each configuration item with
	// the format "key=value", such as "dbName=kusion-db" for type mysql
	Config []string
}

func (o *BackendOptions) AddBackendFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Type, "backend-type", "",
		i18n.T("backend-type specify state storage backend"))
	cmd.Flags().StringSliceVarP(&o.Config, "backend-config", "C", []string{},
		i18n.T("backend-config config state storage backend"))
}

// IsEmpty returns the BackendOptions is empty or not.
func (o *BackendOptions) IsEmpty() bool {
	return o.Type == "" && len(o.Config) == 0
}

// Validate checks the BackendOptions is valid or not.
func (o *BackendOptions) Validate() error {
	if o.Type == "" {
		return ErrEmptyBackendType
	}
	backendFunc := backendinit.GetBackend(o.Type)
	if backendFunc == nil {
		return ErrUnsupportedBackendType
	}
	config, err := o.toStateStorageConfig()
	if err != nil {
		return err
	}
	backendSchema := backendFunc().ConfigSchema()
	if err = validBackendConfig(config, backendSchema); err != nil {
		return err
	}
	return nil
}

// toStateStorageConfig converts BackendOptions to StateStorageConfig.
func (o *BackendOptions) toStateStorageConfig() (*StateStorageConfig, error) {
	config := make(map[string]any)
	for _, v := range o.Config {
		bk := strings.Split(v, "=")
		if len(bk) != 2 {
			return nil, ErrInvalidBackendConfigFormat
		}
		if bk[0] == "" {
			return nil, ErrEmptyBackendConfigKey
		}
		if bk[1] == "" {
			return nil, ErrEmptyBackendConfigValue
		}
		config[bk[0]] = bk[1]
	}

	return &StateStorageConfig{
		Type:   o.Type,
		Config: config,
	}, nil
}

// validBackendConfig checks state backend config from BackendOptions, it only checks whether there are
// unsupported backend configuration items for now.
func validBackendConfig(config *StateStorageConfig, schema cty.Type) error {
	for k := range config.Config {
		if !schema.HasAttribute(k) {
			return fmt.Errorf("%w: %s", ErrUnsupportedBackendConfigItem, k)
		}
	}
	if config.Type == v1.DeprecatedBackendLocal && len(config.Config) != 0 {
		return fmt.Errorf("%w for backend local", ErrNotSupportBackendConfig)
	}
	return nil
}
