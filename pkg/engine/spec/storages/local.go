package storages

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/spec"
)

// LocalStorage should implement the spec.Storage interface.
var _ spec.Storage = &LocalStorage{}

type LocalStorage struct {
	// The path of spec file.
	path string
}

// NewLocalStorage constructs a local filesystem based spec storage.
func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{
		path: path,
	}
}

// Get returns the Spec, if the Spec does not exist, return nil.
func (s *LocalStorage) Get() (*v1.Intent, error) {
	content, err := os.ReadFile(s.path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Empty content, return directly
	if len(content) == 0 {
		return nil, nil
	}

	state := &v1.Intent{}
	err = yaml.Unmarshal(content, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// Apply updates the spec if already exists, or create a new spec.
func (s *LocalStorage) Apply(state *v1.Intent) error {
	if err := os.MkdirAll(filepath.Dir(s.path), os.ModePerm); err != nil {
		fmt.Println(err)
	}

	content, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, content, fs.ModePerm)
}
