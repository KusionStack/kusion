package storages

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	graphstorages "kusionstack.io/kusion/pkg/engine/resource/graph/storages"
	projectstorages "kusionstack.io/kusion/pkg/project/storages"
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

func TestLocalStorage_GraphStorage(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		localStorage       *LocalStorage
		project, workspace string
	}{
		{
			name:    "graph storage from local backend",
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
			mockey.PatchConvey("mock new local graph storage", t, func() {
				mockey.Mock(graphstorages.NewLocalStorage).Return(&graphstorages.LocalStorage{}, nil).Build()
				_, err := tc.localStorage.GraphStorage(tc.project, tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestLocalStorage_ProjectStorage(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		localStorage *LocalStorage
	}{
		{
			name:    "project storage from local backend",
			success: true,
			localStorage: &LocalStorage{
				path: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new local project storage", t, func() {
				mockey.Mock((*projectstorages.LocalStorage).Get).Return(map[string][]string{}, nil).Build()
				_, err := tc.localStorage.ProjectStorage()

				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
