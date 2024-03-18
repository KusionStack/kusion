package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func mockOssStorage() *OssStorage {
	return &OssStorage{bucket: &oss.Bucket{}}
}

func mockOssStorageReadMeta(meta *workspacesMetaData) {
	mockey.Mock((*OssStorage).readMeta).Return(meta, nil).Build()
}

func mockOssStorageWriteMeta() {
	mockey.Mock((*OssStorage).writeMeta).Return(nil).Build()
}

func mockOssStorageWriteWorkspace() {
	mockey.Mock((*OssStorage).writeWorkspace).Return(nil).Build()
}

func TestOssStorage_Get(t *testing.T) {
	testcases := []struct {
		name              string
		success           bool
		wsName            string
		content           []byte
		expectedWorkspace *v1.Workspace
	}{
		{
			name:              "get workspace successfully",
			success:           true,
			wsName:            "dev",
			content:           []byte(mockWorkspaceContent()),
			expectedWorkspace: mockWorkspace("dev"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockey.Mock(oss.Bucket.GetObject).Return(io.NopCloser(bytes.NewReader([]byte(""))), nil).Build()
				mockey.Mock(io.ReadAll).Return(tc.content, nil).Build()
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				workspace, err := mockOssStorage().Get(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedWorkspace, workspace)
			})
		})
	}
}

func TestOssStorage_Create(t *testing.T) {
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
		{
			name:      "failed to create workspace already exist",
			success:   false,
			workspace: mockWorkspace("dev"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				mockOssStorageWriteMeta()
				mockOssStorageWriteWorkspace()
				err := mockOssStorage().Create(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestOssStorage_Update(t *testing.T) {
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
		{
			name:      "failed to update workspace not exist",
			success:   false,
			workspace: mockWorkspace("pre"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				mockOssStorageWriteWorkspace()
				err := mockOssStorage().Update(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestOssStorage_Delete(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		wsName  string
	}{
		{
			name:    "delete workspace successfully",
			success: true,
			wsName:  "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockey.Mock(oss.Bucket.DeleteObject).Return(nil).Build()
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				mockOssStorageWriteMeta()
				err := mockOssStorage().Delete(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestOssStorage_Exist(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		wsName        string
		expectedExist bool
	}{
		{
			name:          "exist workspace",
			success:       true,
			wsName:        "dev",
			expectedExist: true,
		},
		{
			name:          "not exist workspace",
			success:       true,
			wsName:        "pre",
			expectedExist: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				exist, err := mockOssStorage().Exist(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedExist, exist)
			})
		})
	}
}

func TestOssStorage_GetNames(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		expectedNames []string
	}{
		{
			name:          "get all workspace names successfully",
			success:       true,
			expectedNames: []string{"default", "dev", "prod"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				wsNames, err := mockOssStorage().GetNames()
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.Equal(t, tc.expectedNames, wsNames)
				}
			})
		})
	}
}

func TestOssStorage_GetCurrent(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		expectedCurrent string
	}{
		{
			name:            "get current workspace successfully",
			success:         true,
			expectedCurrent: "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				current, err := mockOssStorage().GetCurrent()
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.Equal(t, tc.expectedCurrent, current)
				}
			})
		})
	}
}

func TestOssStorage_SetCurrent(t *testing.T) {
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
		{
			name:    "failed to set current workspace not exist",
			success: false,
			current: "pre",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockOssStorageReadMeta(mockWorkspacesMetaData())
				mockOssStorageWriteMeta()
				err := mockOssStorage().SetCurrent(tc.current)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
