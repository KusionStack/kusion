package storages

import (
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

// LocalStorage is an implementation of backend.Backend which uses local filesystem as storage.
type LocalStorage struct {
	// path is the directory to store the files. If empty, use the default storage path, which depends on
	// the object it's used to store.
	path string
}

func NewLocalStorage(config *v1.BackendLocalConfig) *LocalStorage {
	return &LocalStorage{path: config.Path}
}

func (s *LocalStorage) StateStorage(project, stack, workspace string) state.Storage {
	return statestorages.NewLocalStorage(statestorages.GenStateFilePath(s.path, project, stack, workspace))
}
