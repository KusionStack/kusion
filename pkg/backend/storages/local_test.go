package storages

import (
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
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
		name               string
		localStorage       *LocalStorage
		project, workspace string
		stateStorage       state.Storage
	}{
		{
			name: "state storage from local backend",
			localStorage: &LocalStorage{
				path: "kusion",
			},
			project:   "wordpress",
			workspace: "dev",
			stateStorage: statestorages.NewLocalStorage(
				filepath.Join("kusion", "states", "wordpress", "dev", "state.yaml"),
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stateStorage := tc.localStorage.StateStorage(tc.project, tc.workspace)
			assert.Equal(t, tc.stateStorage, stateStorage)
		})
	}
}

func TestLocalStorage_WorkspaceStorage(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		localStorage *LocalStorage
	}{
		{
			name:    "workspace storage from local backend",
			success: true,
			localStorage: &LocalStorage{
				path: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new local workspace storage", t, func() {
				mockey.Mock(workspacestorages.NewLocalStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
				_, err := tc.localStorage.WorkspaceStorage()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestLocalStorage_ReleaseStorage(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		localStorage       *LocalStorage
		project, workspace string
	}{
		{
			name:    "release storage from local backend",
			success: true,
			localStorage: &LocalStorage{
				path: "kusion",
			},
			project:   "wordpress",
			workspace: "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new local release storage", t, func() {
				mockey.Mock(releasestorages.NewLocalStorage).Return(&releasestorages.LocalStorage{}, nil).Build()
				_, err := tc.localStorage.ReleaseStorage(tc.project, tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
