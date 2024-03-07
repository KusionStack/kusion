package storages

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func TestNewS3Storage(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *v1.BackendS3Config
	}{
		{
			name:    "new S3 storage successfully",
			success: true,
			config: &v1.BackendS3Config{
				GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion",
				},
				Region: "us-east-1",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock s3 session", t, func() {
				mockey.Mock(session.NewSession).Return(&session.Session{}, nil).Build()
				_, err := NewS3Storage(tc.config)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
