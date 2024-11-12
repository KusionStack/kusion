package server

import (
	"github.com/spf13/pflag"
)

type ServerOptions struct {
	Mode               string
	Port               int
	AuthEnabled        bool
	AuthWhitelist      []string
	AuthKeyType        string
	Database           DatabaseOptions
	DefaultBackend     DefaultBackendOptions
	DefaultSource      DefaultSourceOptions
	MaxConcurrent      int
	MaxAsyncConcurrent int
	MaxAsyncBuffer     int
	LogFilePath        string
}

type Options interface {
	// Validate checks Options and return a slice of found error(s)
	Validate() error
	// AddFlags adds flags for a specific Option to the specified FlagSet
	AddFlags(fs *pflag.FlagSet)
}

const (
	MaskString         = "******"
	DefaultPort        = 80
	DefaultAuthKeyType = "RSA"
	DefaultMode        = "KCP"
)
