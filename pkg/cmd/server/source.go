package server

import (
	"net/url"

	"github.com/pkg/errors"
	"kusionstack.io/kusion/pkg/cmd/server/util"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/server"

	"github.com/spf13/pflag"
)

var (
	ErrDefaultSourceRemoteNotSpecified = errors.New("--default-source-remote must be specified")
)

// DefaultSourceOptions holds the default backend access layer configurations.
type DefaultSourceOptions struct {
	SourceRemote string `json:"sourceRemote,omitempty" yaml:"sourceRemote,omitempty"`
}

// ApplyTo uses the run options to read and set the default backend access layer configurations
func (o *DefaultSourceOptions) ApplyTo(config *server.Config) error {
	// populate non-sensitive data
	// convert the source remote string to a URL
	sourceRemote, err := url.Parse(o.SourceRemote)
	if err != nil {
		return err
	}
	config.DefaultSource.Remote = sourceRemote
	config.DefaultSource.SourceProvider = constant.DefaultSourceType
	config.DefaultSource.Description = constant.DefaultSourceDesc
	return nil
}

// Validate checks validation of DefaultSourceOptions
func (o *DefaultSourceOptions) Validate() error {
	var errs []error
	if o.SourceRemote == "" {
		errs = append(errs, ErrDefaultSourceRemoteNotSpecified)
	}
	if errs != nil {
		err := util.AggregateError(errs)
		return errors.Wrap(err, "invalid source options")
	}
	return nil
}

// AddFlags adds flags related to backend to a specified FlagSet
func (o *DefaultSourceOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SourceRemote, "default-source-remote", o.SourceRemote, "the default source remote")
}
