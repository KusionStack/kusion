package storages

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	releasestorages "kusionstack.io/kusion/pkg/engine/release/storages"
	graphstorages "kusionstack.io/kusion/pkg/engine/resource/graph/storages"
	projectstorages "kusionstack.io/kusion/pkg/project/storages"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
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

func TestS3Storage_WorkspaceStorage(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		s3Storage *S3Storage
	}{
		{
			name:    "workspace storage from s3 backend",
			success: true,
			s3Storage: &S3Storage{
				s3:     &s3.S3{},
				bucket: "infra",
				prefix: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new s3 workspace storage", t, func() {
				mockey.Mock(workspacestorages.NewS3Storage).Return(&workspacestorages.S3Storage{}, nil).Build()
				_, err := tc.s3Storage.WorkspaceStorage()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_ReleaseStorage(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		s3Storage          *S3Storage
		project, workspace string
	}{
		{
			name:    "release storage from s3 backend",
			success: true,
			s3Storage: &S3Storage{
				s3:     &s3.S3{},
				bucket: "infra",
				prefix: "kusion",
			},
			project:   "wordpress",
			workspace: "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new s3 release storage", t, func() {
				mockey.Mock(releasestorages.NewS3Storage).Return(&releasestorages.S3Storage{}, nil).Build()
				_, err := tc.s3Storage.ReleaseStorage(tc.project, tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_GraphStorage(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		s3Storage          *S3Storage
		project, workspace string
	}{
		{
			name:    "graph storage from s3 backend",
			success: true,
			s3Storage: &S3Storage{
				s3:     &s3.S3{},
				bucket: "infra",
				prefix: "kusion",
			},
			project:   "wordpress",
			workspace: "dev",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new s3 graph storage", t, func() {
				mockey.Mock(graphstorages.NewS3Storage).Return(&graphstorages.S3Storage{}, nil).Build()
				_, err := tc.s3Storage.GraphStorage(tc.project, tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestS3Storage_ProjectStorage(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		s3Storage *S3Storage
	}{
		{
			name:    "project storage from s3 backend",
			success: true,
			s3Storage: &S3Storage{
				s3:     &s3.S3{},
				bucket: "infra",
				prefix: "kusion",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock new s3 project storage", t, func() {
				mockey.Mock((*projectstorages.S3Storage).Get).Return(map[string][]string{}, nil).Build()
				_, err := tc.s3Storage.ProjectStorage()

				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
