package backend

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	backendInit "kusionstack.io/kusion/pkg/engine/backend/init"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/util/i18n"
)

// backend config state storage type
type Storage struct {
	Type   string                 `json:"storageType,omitempty" yaml:"storageType,omitempty"`
	Config map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// BackendOps kusion cli backend override config
type BackendOps struct {
	// Config is a series of backend configurations,
	// such as ["path=kusion_state.json"]
	Config []string

	// Type is the type of backend, currently supported:
	//    local - state is stored to a local file
	//    db 	- state is stored to db
	//    oss 	- state is stored to aliyun oss
	//    s3 	- state is stored to aws s3
	Type string
}

func (o *BackendOps) AddBackendFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Type, "backend-type", "",
		i18n.T("backend-type specify state storage backend"))
	cmd.Flags().StringSliceVarP(&o.Config, "backend-config", "C", []string{},
		i18n.T("backend-config config state storage backend"))
}

// MergeConfig merge project backend config and cli backend config
func MergeConfig(config, override map[string]interface{}) map[string]interface{} {
	content := make(map[string]interface{})
	for k, v := range config {
		content[k] = v
	}
	for k, v := range override {
		content[k] = v
	}
	return content
}

// NewDefaultBackend return default backend, default backend is local filesystem
func NewDefaultBackend(dir string, fileName string) *Storage {
	return &Storage{
		Type: "local",
		Config: map[string]interface{}{
			"path": filepath.Join(dir, fileName),
		},
	}
}

// BackendFromConfig return stateStorage, this func handler
// backend config merge and configure backend.
// return a StateStorage to manage State
func BackendFromConfig(config *Storage, override BackendOps, dir string) (states.StateStorage, error) {
	var backendConfig Storage
	if config == nil {
		config = NewDefaultBackend(dir, local.KusionState)
	}
	if config.Type != "" {
		backendConfig.Type = config.Type
	}

	if override.Type != "" {
		backendConfig.Type = override.Type
	}
	configOverride := make(map[string]interface{})
	for _, v := range override.Config {
		bk := strings.Split(v, "=")
		if len(bk) != 2 {
			return nil, fmt.Errorf("kusion cli backend config should be path=kusion_state.json")
		}
		configOverride[bk[0]] = bk[1]
	}
	if config.Config != nil || override.Config != nil {
		backendConfig.Config = MergeConfig(config.Config, configOverride)
	}

	backendFunc := backendInit.GetBackend(backendConfig.Type)
	if backendFunc == nil {
		return nil, fmt.Errorf("kusion backend storage: %s not support, please check storageType config", backendConfig.Type)
	}

	bf := backendFunc()

	backendSchema := bf.ConfigSchema()
	err := validBackendConfig(backendConfig.Config, backendSchema)
	if err != nil {
		return nil, err
	}
	ctyBackend, err := gocty.ToCtyValue(backendConfig.Config, backendSchema)
	if err != nil {
		return nil, err
	}

	err = bf.Configure(ctyBackend)
	if err != nil {
		return nil, err
	}

	return bf.StateStorage(), nil
}

// validBackendConfig check backend config.
func validBackendConfig(config map[string]interface{}, schema cty.Type) error {
	for k := range config {
		if !schema.HasAttribute(k) {
			return fmt.Errorf("not support %s in backend config", k)
		}
	}
	return nil
}
