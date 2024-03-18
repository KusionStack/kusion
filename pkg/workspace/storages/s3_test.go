package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func mockS3Storage() *S3Storage {
	return &S3Storage{s3: &s3.S3{}}
}

func mockS3StorageReadMeta(meta *workspacesMetaData) {
	mockey.Mock((*S3Storage).readMeta).Return(meta, nil).Build()
}

func mockS3StorageWriteMeta() {
	mockey.Mock((*S3Storage).writeMeta).Return(nil).Build()
}

func mockS3StorageWriteWorkspace() {
	mockey.Mock((*S3Storage).writeWorkspace).Return(nil).Build()
}

func TestS3Storage_Get(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockey.Mock((*s3.S3).GetObject).Return(&s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(""))),
				}, nil).Build()
				mockey.Mock(io.ReadAll).Return(tc.content, nil).Build()
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				workspace, err := mockS3Storage().Get(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedWorkspace, workspace)
			})
		})
	}
}

func TestS3Storage_Create(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				mockS3StorageWriteMeta()
				mockS3StorageWriteWorkspace()
				err := mockS3Storage().Create(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_Update(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				mockS3StorageWriteWorkspace()
				err := mockS3Storage().Update(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_Delete(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockey.Mock((*s3.S3).DeleteObject).Return(nil, nil).Build()
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				mockS3StorageWriteMeta()
				err := mockS3Storage().Delete(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_Exist(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				exist, err := mockS3Storage().Exist(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.expectedExist, exist)
			})
		})
	}
}

func TestS3Storage_GetNames(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				wsNames, err := mockS3Storage().GetNames()
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.Equal(t, tc.expectedNames, wsNames)
				}
			})
		})
	}
}

func TestS3Storage_GetCurrent(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				current, err := mockS3Storage().GetCurrent()
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.Equal(t, tc.expectedCurrent, current)
				}
			})
		})
	}
}

func TestS3Storage_SetCurrent(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockS3StorageReadMeta(mockWorkspacesMetaData())
				mockS3StorageWriteMeta()
				err := mockS3Storage().SetCurrent(tc.current)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
