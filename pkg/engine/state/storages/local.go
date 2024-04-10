package storages

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// LocalStorage is an implementation of state.Storage which uses local filesystem as storage.
type LocalStorage struct {
	// The path of state file.
	path string
}

func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{path: path}
}

func (s *LocalStorage) Get() (*v1.DeprecatedState, error) {
	content, err := os.ReadFile(s.path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(content) != 0 {
		state := &v1.DeprecatedState{}
		err = yaml.Unmarshal(content, state)
		if err != nil {
			return nil, err
		}
		return state, nil
	} else {
		return nil, nil
	}
}

func (s *LocalStorage) Apply(state *v1.DeprecatedState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), os.ModePerm); err != nil {
		fmt.Println(err)
	}

	content, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, content, fs.ModePerm)
}
