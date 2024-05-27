package storages

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
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
