package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/util/kfile"
)

func testDataFolder() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "testdata")
}

func mockValidOperator() *Operator {
	return &Operator{
		storagePath: filepath.Join(testDataFolder(), defaultRelativeStoragePath),
	}
}

func TestNewDefaultOperator(t *testing.T) {
	t.Run("new default operator successfully", func(t *testing.T) {
		mockey.PatchConvey("mock kusion data folder", t, func() {
			mockey.Mock(kfile.KusionDataFolder).Return(testDataFolder(), nil).Build()

			operator, err := NewDefaultOperator()
			storagePath := filepath.Join(testDataFolder(), defaultRelativeStoragePath)
			assert.Nil(t, err)
			assert.Equal(t, storagePath, operator.storagePath)
			assert.DirExists(t, storagePath)
		})
	})
}

func TestOperator_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		success  bool
		operator *Operator
	}{
		{
			name:     "valid operator",
			success:  true,
			operator: mockValidOperator(),
		},
		{
			name:     "invalid operator empty storage path",
			success:  false,
			operator: &Operator{},
		},
		{
			name:    "invalid operator not yaml workspace",
			success: false,
			operator: &Operator{
				storagePath: filepath.Join(testDataFolder(), "invalid_workspaces_not_yaml"),
			},
		},
		{
			name:    "invalid operator dir workspace",
			success: false,
			operator: &Operator{
				storagePath: filepath.Join(testDataFolder(), "invalid_workspaces_dir"),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operator.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestOperator_GetWorkspaceNames(t *testing.T) {
	operator, _ := NewOperator(filepath.Join(testDataFolder(), "workspaces_for_list"))
	testcases := []struct {
		name     string
		success  bool
		operator *Operator
		wsNames  []string
	}{
		{
			name:     "get workspace successfully",
			success:  true,
			operator: operator,
			wsNames:  []string{"dev", "pre", "prod"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			wsNames, err := tc.operator.GetWorkspaceNames()
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				assert.Equal(t, tc.wsNames, wsNames)
			}
		})
	}
}

func TestOperator_GetWorkspace(t *testing.T) {
	testcases := []struct {
		name     string
		success  bool
		operator *Operator
		wsName   string
	}{
		{
			name:     "get workspace successfully",
			success:  true,
			operator: mockValidOperator(),
			wsName:   "for_get_ws",
		},
		{
			name:     "failed to get workspace not exist",
			success:  false,
			operator: mockValidOperator(),
			wsName:   "for_get_failure_ws",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ws, err := tc.operator.GetWorkspace(tc.wsName)
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				assert.Equal(t, mockValidWorkspace(tc.wsName), ws)
			}
		})
	}
}

func TestOperator_CreateWorkspace(t *testing.T) {
	testcases := []struct {
		name     string
		success  bool
		operator *Operator
		ws       *v1.Workspace
	}{
		{
			name:     "create workspace successfully",
			success:  true,
			operator: mockValidOperator(),
			ws:       mockValidWorkspace("for_create_ws"),
		},
		{
			name:     "failed to create workspace already exists",
			success:  false,
			operator: mockValidOperator(),
			ws:       mockValidWorkspace("for_create_failure_ws"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operator.CreateWorkspace(tc.ws)
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				_ = tc.operator.DeleteWorkspace(tc.ws.Name)
			}
		})
	}
}

func TestOperator_UpdateWorkspace(t *testing.T) {
	testcases := []struct {
		name     string
		success  bool
		operator *Operator
		ws       *v1.Workspace
	}{
		{
			name:     "update workspace successfully",
			success:  true,
			operator: mockValidOperator(),
			ws:       mockValidWorkspace("for_update_ws"),
		},
		{
			name:     "failed to update workspace not exist",
			success:  false,
			operator: mockValidOperator(),
			ws:       mockValidWorkspace("for_update_failure_ws"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operator.UpdateWorkspace(tc.ws)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestOperator_DeleteWorkspace(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		operator     *Operator
		wsName       string
		needRecreate bool
	}{
		{
			name:         "delete workspace successfully",
			success:      true,
			operator:     mockValidOperator(),
			wsName:       "for_delete_ws",
			needRecreate: true,
		},
		{
			name:         "delete workspace not exist successfully",
			success:      true,
			operator:     mockValidOperator(),
			wsName:       "for_delete_not_exist_ws",
			needRecreate: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operator.DeleteWorkspace(tc.wsName)
			assert.Equal(t, tc.success, err == nil)
			if err == nil && tc.needRecreate {
				ws := mockValidWorkspace(tc.wsName)
				_ = tc.operator.CreateWorkspace(ws)
			}
		})
	}
}
