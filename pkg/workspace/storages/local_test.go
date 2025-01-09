package storages

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func testDataFolder(path string) string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "testdata", path)
}

func mockWorkspace(name string) *v1.Workspace {
	return &v1.Workspace{
		Name: name,
		Modules: map[string]*v1.ModuleConfig{
			"mysql": {
				Path:    "ghcr.io/kusionstack/mysql",
				Version: "0.1.0",
				Configs: v1.Configs{
					Default: v1.GenericConfig{
						"type":         "aws",
						"version":      "5.7",
						"instanceType": "db.t3.micro",
					},
					ModulePatcherConfigs: v1.ModulePatcherConfigs{
						"smallClass": {
							GenericConfig: v1.GenericConfig{
								"instanceType": "db.t3.small",
							},
							ProjectSelector: []string{"foo", "bar"},
						},
					},
				},
			},
			"network": {
				Configs: v1.Configs{
					Default: v1.GenericConfig{
						"type": "aws",
					},
				},
			},
		},
		Context: map[string]any{
			"kubernetes": v1.GenericConfig{
				"config": "/etc/kubeconfig.yaml",
			},
		},
	}
}

func mockWorkspaceContent() string {
	return `
modules:
  mysql:
    path: ghcr.io/kusionstack/mysql
    version: 0.1.0
    configs:
      default:
        instanceType: db.t3.micro
        type: aws
        version: '5.7'
      smallClass:
        projectSelector:
          - foo
          - bar
        instanceType: db.t3.small
  network:
    configs:
      default:
        type: aws
context:
    kubernetes:
        config: /etc/kubeconfig.yaml
`
}

func TestLocalStorageOperation(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		path         string
		expectedMeta *workspacesMetaData
		deletePath   bool
	}{
		{
			name:    "new local storage with empty directory",
			success: true,
			path:    testDataFolder("empty_workspaces"),
			expectedMeta: &workspacesMetaData{
				Current:             "default",
				AvailableWorkspaces: []string{"default"},
			},
			deletePath: true,
		},
		{
			name:         "new local storage with exist directory",
			success:      true,
			path:         testDataFolder("workspaces"),
			expectedMeta: mockWorkspacesMetaData(),
			deletePath:   false,
		},
		{
			name:         "new local storage failed",
			success:      false,
			path:         testDataFolder("invalid_metadata_workspaces"),
			expectedMeta: nil,
			deletePath:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(tc.path)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedMeta, s.meta)
			}
			if tc.deletePath {
				_ = os.RemoveAll(tc.path)
			}
		})
	}
}

func TestLocalStorage_Get(t *testing.T) {
	testcases := []struct {
		name              string
		success           bool
		wsName            string
		expectedWorkspace *v1.Workspace
	}{
		{
			name:              "get workspace successfully",
			success:           true,
			wsName:            "dev",
			expectedWorkspace: mockWorkspace("dev"),
		},
		{
			name:              "get workspace failed not exist",
			success:           false,
			wsName:            "pre",
			expectedWorkspace: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("workspaces"))
			assert.NoError(t, err)
			workspace, err := s.Get(tc.wsName)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedWorkspace, workspace)
			}
		})
	}
}

func TestLocalStorage_Create(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		path         string
		workspace    *v1.Workspace
		expectedMeta *workspacesMetaData
	}{
		{
			name:      "create workspace successfully",
			success:   true,
			path:      testDataFolder("for_create_workspaces"),
			workspace: mockWorkspace("dev"),
			expectedMeta: &workspacesMetaData{
				Current:             "default",
				AvailableWorkspaces: []string{"default", "dev"},
			},
		},
		{
			name:         "create workspace failed already exist",
			success:      false,
			path:         testDataFolder("workspaces"),
			workspace:    mockWorkspace("prod"),
			expectedMeta: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(tc.path)
			assert.NoError(t, err)
			err = s.Create(tc.workspace)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedMeta, s.meta)
				_ = s.Delete(tc.workspace.Name)
			}
		})
	}
}

func TestLocalStorage_Update(t *testing.T) {
	testcases := []struct {
		name              string
		success           bool
		workspace         *v1.Workspace
		expectedWorkspace *v1.Workspace
	}{
		{
			name:              "update workspace successfully",
			success:           true,
			workspace:         &v1.Workspace{Name: "default"},
			expectedWorkspace: &v1.Workspace{Name: "default"},
		},
		{
			name:              "update workspace failed not exist",
			success:           false,
			workspace:         mockWorkspace("pre"),
			expectedWorkspace: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("workspaces"))
			assert.NoError(t, err)
			err = s.Update(tc.workspace)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				var workspace *v1.Workspace
				workspace, err = s.Get(tc.workspace.Name)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedWorkspace, workspace)
			}
		})
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		path         string
		wsName       string
		expectedMeta *workspacesMetaData
	}{
		{
			name:    "delete workspace successfully",
			success: true,
			path:    testDataFolder("for_delete_workspaces"),
			wsName:  "dev",
			expectedMeta: &workspacesMetaData{
				Current:             "default",
				AvailableWorkspaces: []string{"default"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(tc.path)
			assert.NoError(t, err)
			err = s.Delete(tc.wsName)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedMeta, s.meta)
				_ = s.Create(mockWorkspace(tc.wsName))
			}
		})
	}
}

func TestLocalStorage_GetNames(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		expectedNames []string
	}{
		{
			name:          "exist workspace",
			success:       true,
			expectedNames: []string{"default", "dev", "prod"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("workspaces"))
			assert.NoError(t, err)
			names, err := s.GetNames()
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedNames, names)
			}
		})
	}
}

func TestLocalStorage_GetCurrent(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		expectedCurrent string
	}{
		{
			name:            "current default workspace",
			success:         true,
			expectedCurrent: "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("workspaces"))
			assert.NoError(t, err)
			current, err := s.GetCurrent()
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.Equal(t, tc.expectedCurrent, current)
			}
		})
	}
}

func TestLocalStorage_SetCurrent(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		wsName  string
	}{
		{
			name:    "set current workspace successfully",
			success: true,
			wsName:  "dev",
		},
		{
			name:    "failed to set current workspace not exist",
			success: false,
			wsName:  "pre",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(testDataFolder("for_set_current_workspaces"))
			assert.NoError(t, err)
			err = s.SetCurrent(tc.wsName)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				var current string
				current, err = s.GetCurrent()
				assert.NoError(t, err)
				assert.Equal(t, tc.wsName, current)
				_ = s.SetCurrent("default")
			}
		})
	}
}

func TestLocalStorage_RenameWorkspace(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		oldName string
		newName string
	}{
		{
			name:    "failed to rename workspace name is empty",
			success: false,
			oldName: "dev",
			newName: "",
		},
		{
			name:    "rename workspace successfully",
			success: true,
			oldName: "dev",
			newName: "newName",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.RemoveAll(testDataFolder("for_rename_workspaces"))
			s, err := NewLocalStorage(testDataFolder("for_rename_workspaces"))
			assert.NoError(t, err)
			err = s.Create(mockWorkspace(tc.oldName))
			assert.NoError(t, err)
			err = s.RenameWorkspace(tc.oldName, tc.newName)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				names, err := s.GetNames()
				assert.NoError(t, err)
				assert.Contains(t, names, tc.newName)
				assert.NotContains(t, names, tc.oldName)
			}
		})
	}
}
