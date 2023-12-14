package local

import (
	"errors"

	"github.com/zclconf/go-cty/cty"

	"kusionstack.io/kusion/pkg/engine/states"
)

type LocalBackend struct {
	FileSystemState
}

func NewLocalBackend() states.Backend {
	return &LocalBackend{}
}

func (f *LocalBackend) StateStorage() states.StateStorage {
	return &FileSystemState{f.Path}
}

func (f *LocalBackend) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"path": cty.String,
	}
	return cty.Object(config)
}

func (f *LocalBackend) Configure(obj cty.Value) error {
	var path cty.Value
	// path should be configured by kusion, not by workspace or cli flags.
	if path = obj.GetAttr("path"); path.IsNull() || path.AsString() == "" {
		return errors.New("path must be configure in backend config")
	}
	f.Path = path.AsString()
	return nil
}
