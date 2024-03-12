package storages

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/engine/state"
	statestorages "kusionstack.io/kusion/pkg/engine/state/storages"
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
				mockey.Mock(s3.New).Return(&s3.S3{}).Build()
				_, err := NewS3Storage(tc.config)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_StateStorage(t *testing.T) {
	testcases := []struct {
		name                      string
		s3Storage                 *S3Storage
		project, stack, workspace string
		stateStorage              state.Storage
	}{
		{
			name: "state storage from s3 backend",
			s3Storage: &S3Storage{
				s3:     &s3.S3{},
				bucket: "infra",
				prefix: "kusion",
			},
			project:   "wordpress",
			stack:     "dev",
			workspace: "dev",
			stateStorage: statestorages.NewS3Storage(
				&s3.S3{},
				"infra",
				"kusion/states/wordpress/dev/dev/state.yaml",
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stateStorage := tc.s3Storage.StateStorage(tc.project, tc.stack, tc.workspace)
			assert.Equal(t, tc.stateStorage, stateStorage)
		})
	}
}
