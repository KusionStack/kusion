package storages

import (
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
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
