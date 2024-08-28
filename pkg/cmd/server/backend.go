package server

import (
	"os"

	"github.com/pkg/errors"
	"kusionstack.io/kusion/pkg/server"

	"github.com/spf13/pflag"
)

var (
	ErrBackendTypeNotSpecified     = errors.New("--default-backend-type must be specified")
	ErrBackendNameNotSpecified     = errors.New("--default-backend-name must be specified")
	ErrBackendEndpointNotSpecified = errors.New("--default-backend-endpoint must be specified")
	ErrBackendAKSKNotSpecified     = errors.New("no AK/SK found. Please set OSS_ACCESS_KEY_ID and OSS_ACCESS_KEY_SECRET environment variables")
)

// DefaultBackendOptions holds the default backend access layer configurations.
type DefaultBackendOptions struct {
	BackendName     string `json:"backendName,omitempty" yaml:"backendName,omitempty"`
	BackendType     string `json:"backendType,omitempty" yaml:"backendType,omitempty"`
	BackendEndpoint string `json:"backendEndpoint,omitempty" yaml:"backendEndpoint,omitempty"`
}

// ApplyTo uses the run options to read and set the default backend access layer configurations
func (o *DefaultBackendOptions) ApplyTo(config *server.Config) error {
	// populate non-sensitive data
	config.DefaultBackend.BackendConfig.Type = o.BackendType
	config.DefaultBackend.BackendConfig.Configs = make(map[string]any)
	config.DefaultBackend.BackendConfig.Configs["bucket"] = o.BackendName
	config.DefaultBackend.BackendConfig.Configs["endpoint"] = o.BackendEndpoint

	// populate sensitive data
	backendAccessKey := os.Getenv("BACKEND_ACCESS_KEY_ID")
	backendAccessSecret := os.Getenv("BACKEND_ACCESS_KEY_SECRET")
	config.DefaultBackend.BackendConfig.Configs["accessKeyID"] = backendAccessKey
	config.DefaultBackend.BackendConfig.Configs["accessKeySecret"] = backendAccessSecret
	return nil
}

// AddFlags adds flags related to backend to a specified FlagSet
func (o *DefaultBackendOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BackendEndpoint, "default-backend-endpoint", o.BackendEndpoint, "the default backend endpoint")
	fs.StringVar(&o.BackendName, "default-backend-name", o.BackendName, "the default backend name")
	fs.StringVar(&o.BackendType, "default-backend-type", o.BackendType, "the default backend type")
}
