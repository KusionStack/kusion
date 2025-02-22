package storages

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockS3Storage(meta *workspacesMetaData) *S3Storage {
	return &S3Storage{s3: &s3.S3{}, meta: meta}
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
				workspace, err := mockS3Storage(mockWorkspacesMetaData()).Get(tc.wsName)
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
				mockS3StorageWriteMeta()
				mockS3StorageWriteWorkspace()
				err := mockS3Storage(mockWorkspacesMetaData()).Create(tc.workspace)
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
				mockS3StorageWriteWorkspace()
				err := mockS3Storage(mockWorkspacesMetaData()).Update(tc.workspace)
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
				mockS3StorageWriteMeta()
				err := mockS3Storage(mockWorkspacesMetaData()).Delete(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
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
				wsNames, err := mockS3Storage(mockWorkspacesMetaData()).GetNames()
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
				current, err := mockS3Storage(mockWorkspacesMetaData()).GetCurrent()
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
				mockS3StorageWriteMeta()
				err := mockS3Storage(mockWorkspacesMetaData()).SetCurrent(tc.current)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_RenameWorkspace(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		oldName string
		newName string
	}{
		{
			name:    "rename workspace successfully",
			success: true,
			oldName: "dev",
			newName: "newName",
		},
		{
			name:    "failed to rename workspace name is empty",
			success: false,
			oldName: "",
			newName: "newName",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock s3 operation", t, func() {
				mockey.Mock((*s3.S3).CopyObject).Return(&s3.CopyObjectOutput{}, nil).Build()
				mockey.Mock((*s3.S3).PutObject).Return(&s3.PutObjectOutput{}, nil).Build()
				mockey.Mock((*s3.S3).DeleteObject).Return(&s3.DeleteObjectOutput{}, nil).Build()
				err := mockS3Storage(mockWorkspacesMetaData()).RenameWorkspace(tc.oldName, tc.newName)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
