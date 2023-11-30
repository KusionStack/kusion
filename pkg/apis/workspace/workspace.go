package workspace

import (
	"errors"
	"reflect"
)

var ErrEmptyWorkspaceName = errors.New("empty workspace name")

// Workspace is a logical concept representing a target that stacks will be deployed to.
// Workspace is managed by platform engineers, which contains a set of configurations
// that application developers do not want or should not concern, and is reused by multiple
// stacks belonging to different projects.
type Workspace struct {
	// Name identifies a Workspace uniquely.
	Name string `yaml:"-" json:"-"`

	// Modules are the configs of a set of modules.
	Modules ModuleConfigs `yaml:"modules,omitempty" json:"modules,omitempty"`

	// Runtimes are the configs of a set of runtimes.
	Runtimes RuntimeConfigs `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`

	// Backends are the configs of a set of backends.
	Backends BackendConfigs `yaml:"backends,omitempty" json:"backends,omitempty"`
}

// Validate is used to validate the Workspace is valid or not.
func (w Workspace) Validate() error {
	if w.Name == "" {
		return ErrEmptyWorkspaceName
	}
	if !reflect.DeepEqual(w.Modules, ModuleConfigs{}) {
		if err := w.Modules.Validate(); err != nil {
			return err
		}
	}
	if !reflect.DeepEqual(w.Runtimes, RuntimeConfigs{}) {
		if err := w.Runtimes.Validate(); err != nil {
			return err
		}
	}
	if !reflect.DeepEqual(w.Backends, BackendConfigs{}) {
		if err := w.Modules.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GenericConfig is a generic model to describe config which shields the difference among multiple concrete
// models. GenericConfig is designed for extensibility, used for module, terraform runtime config, etc.
type GenericConfig map[string]any
