package storages

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
)

func TestNewLocalStorage(t *testing.T) {
	testcases := []struct {
		name    string
		config  *v1.BackendLocalConfig
		storage *LocalStorage
	}{
		{
			name:    "new local storage successfully",
			config:  &v1.BackendLocalConfig{Path: "etc"},
			storage: &LocalStorage{path: "etc"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewLocalStorage(tc.config)
			assert.Equal(t, tc.storage, storage)
		})
	}
}

func TestLocalStorage_StateStorage(t *testing.T) {
	testcases := []struct {
		name                      string
		localStorage              *LocalStorage
		project, stack, workspace string
		stateStorage              state.Storage
	}{
		{
			name: "state storage from s3 backend",
			localStorage: &LocalStorage{
				path: "kusion",
			},
			project:   "wordpress",
			stack:     "dev",
			workspace: "dev",
			stateStorage: statestorages.NewLocalStorage(
				filepath.Join("kusion", "states", "wordpress", "dev", "dev", "state.yaml"),
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stateStorage := tc.localStorage.StateStorage(tc.project, tc.stack, tc.workspace)
			assert.Equal(t, stateStorage, tc.stateStorage)
		})
	}
}
