package storages

import (
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

func TestNewOssStorage(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendOssConfig
	}{
		{
			name:    "new oss storage successfully",
			success: true,
			config: &v1.BackendOssConfig{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss client", t, func() {
				mockey.Mock(oss.New).Return(&oss.Client{}, nil).Build()
				_, err := NewOssStorage(tc.config)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestOssStorage_WorkspaceStorage(t *testing.T) {
	testcases := []struct {
		name       string
		success    bool
		ossStorage *OssStorage
	}{
		{
			name:    "workspace storage from oss backend",
			success: true,
			ossStorage: &OssStorage{
				bucket: &oss.Bucket{},
				prefix: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new oss workspace storage", t, func() {
				mockey.Mock(workspacestorages.NewOssStorage).Return(&workspacestorages.OssStorage{}, nil).Build()
				_, err := tc.ossStorage.WorkspaceStorage()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestOssStorage_ReleaseStorage(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		ossStorage         *OssStorage
		project, workspace string
	}{
		{
			name:    "release storage from s3 backend",
			success: true,
			ossStorage: &OssStorage{
				bucket: &oss.Bucket{},
				prefix: "kusion",
			},
			project:   "wordpress",
			workspace: "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new oss release storage", t, func() {
				mockey.Mock(releasestorages.NewOssStorage).Return(&releasestorages.OssStorage{}, nil).Build()
				_, err := tc.ossStorage.ReleaseStorage(tc.project, tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
