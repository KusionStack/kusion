package workspace

import (
	"errors"
	"reflect"
)

var ErrEmptyLocalFilePath = errors.New("empty local file path")

// BackendConfigs contains config of the backend, which is used to store state, etc. Only one kind
// backend can be configured.
// todo: add more backends declared in pkg/engine/backend
type BackendConfigs struct {
	// Local is backend to use local file system.
	Local LocalFileConfig `yaml:"local,omitempty" json:"local,omitempty"`
}

// Validate is used to validate BackendConfigs is valid or not.
func (b *BackendConfigs) Validate() error {
	if !reflect.DeepEqual(b.Local, LocalFileConfig{}) {
		if err := b.Local.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// LocalFileConfig contains the config of using local file system as backend.
type LocalFileConfig struct {
	// Path is place to store state, etc.
	Path string `yaml:"path" json:"path"`
}

// Validate is used to validate LocalFileConfig is valid or not.
func (b *LocalFileConfig) Validate() error {
	if b.Path == "" {
		return ErrEmptyLocalFilePath
	}
	return nil
}
