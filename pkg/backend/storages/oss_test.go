package storages

import (
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
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

func TestOssStorage_StateStorage(t *testing.T) {
	testcases := []struct {
		name                      string
		ossStorage                *OssStorage
		project, stack, workspace string
		stateStorage              state.Storage
	}{
		{
			name: "state storage from oss backend",
			ossStorage: &OssStorage{
				bucket: &oss.Bucket{},
				prefix: "kusion",
			},
			project:   "wordpress",
			stack:     "dev",
			workspace: "dev",
			stateStorage: statestorages.NewOssStorage(
				&oss.Bucket{},
				"kusion/states/wordpress/dev/dev/state.yaml",
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stateStorage := tc.ossStorage.StateStorage(tc.project, tc.stack, tc.workspace)
			assert.Equal(t, stateStorage, tc.stateStorage)
		})
	}
}
