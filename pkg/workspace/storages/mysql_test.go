package storages

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func mockMysqlStorage() *MysqlStorage {
	return &MysqlStorage{db: &gorm.DB{}}
}

func TestMysqlStorage_Get(t *testing.T) {
	testcases := []struct {
		name              string
		success           bool
		wsName            string
		content           string
		expectedWorkspace *v1.Workspace
	}{
		{
			name:              "get workspace successfully",
			success:           true,
			wsName:            "dev",
			content:           mockWorkspaceContent(),
			expectedWorkspace: mockWorkspace("dev"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock gorm operation", t, func() {
				mockey.Mock(checkWorkspaceExistenceInMysql).Return(true, nil).Build()
				mockey.Mock(getWorkspaceFromMysql).Return(&WorkspaceMysqlDO{Content: tc.content}, nil).Build()
				workspace, err := mockMysqlStorage().Get(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedWorkspace, workspace)
			})
		})
	}
}

func TestMysqlStorage_Create(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		workspace *v1.Workspace
	}{
		{
			name:      "create workspace successfully",
			success:   true,
			workspace: mockWorkspace("pre"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock gorm operation", t, func() {
				mockey.Mock(checkWorkspaceExistenceInMysql).Return(false, nil).Build()
				mockey.Mock(createWorkspaceInMysql).Return(nil).Build()
				err := mockMysqlStorage().Create(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestMysqlStorage_Update(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		workspace *v1.Workspace
	}{
		{
			name:      "update workspace successfully",
			success:   true,
			workspace: mockWorkspace("dev"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock gorm operation", t, func() {
				mockey.Mock(checkWorkspaceExistenceInMysql).Return(true, nil).Build()
				mockey.Mock(updateWorkspaceInMysql).Return(nil).Build()
				err := mockMysqlStorage().Update(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestMysqlStorage_SetCurrent(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		current string
	}{
		{
			name:    "set current workspace successfully",
			success: true,
			current: "prod",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock gorm operation", t, func() {
				mockey.Mock(checkWorkspaceExistenceInMysql).Return(true, nil).Build()
				mockey.Mock(alterCurrentWorkspaceInMysql).Return(nil).Build()
				err := mockMysqlStorage().SetCurrent(tc.current)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
