package local

import (
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
	if path = obj.GetAttr("path"); !path.IsNull() && path.AsString() != "" {
		f.Path = path.AsString()
	} else {
		f.Path = KusionState
	}
	return nil
}
