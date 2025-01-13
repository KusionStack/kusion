package storages

import (
	"context"
	"testing"

	googlestorage "cloud.google.com/go/storage"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	googleauth "golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockGoogleBucketHandle() *googlestorage.BucketHandle {
	config := &v1.BackendGoogleConfig{
		Credentials: &googleauth.Credentials{
			JSON: []byte(`{
                "type": "service_account",
                "project_id": "project-id",
                "private_key_id": "private-key-id",
                "private_key": "private_key",
                "client_email": "client-email",
                "client_id": "client-id",
                "auth_uri": "auth-uri",
                "token_uri": "token-uri",
                "auth_provider_x509_cert_url": "auth-provider-x509-cert-url",
                "client_x509_cert_url": "client-x509-cert-url"
            }`),
		},
		GenericBackendObjectStorageConfig: &v1.GenericBackendObjectStorageConfig{
			Bucket: "valid-bucket",
			Prefix: "valid-prefix",
		},
	}
	client, err := googlestorage.NewClient(context.Background(), option.WithCredentials(config.Credentials))
	if err != nil {
		return nil
	}
	bucket := client.Bucket(config.Bucket)
	return bucket
}

func mockGoogleStorage() *GoogleStorage {
	return &GoogleStorage{
		bucket: *mockGoogleBucketHandle(),
		prefix: "valid-prefix",
		meta:   mockReleasesMeta(),
	}
}

func mockGoogleStorageWriteMeta() {
	mockey.Mock((*GoogleStorage).writeMeta).Return(nil).Build()
}

func mockGoogleStorageWriteRelease() {
	mockey.Mock((*GoogleStorage).writeRelease).Return(nil).Build()
}

func TestNewGoogleStorage(t *testing.T) {
	tests := []struct {
		name    string
		bucket  *googlestorage.BucketHandle
		prefix  string
		wantErr bool
	}{
		{
			name:    "valid bucket and prefix",
			bucket:  mockGoogleBucketHandle(),
			prefix:  "valid-prefix",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockey.Mock((*GoogleStorage).readMeta).Return(nil).Build()
				got, err := NewGoogleStorage(tt.bucket, tt.prefix)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewGoogleStorage() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					assert.NotNil(t, got)
					assert.Equal(t, tt.prefix, got.prefix)
				}
			})
		})
	}
}

func TestGoogleStorage_GetRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		expectedRevisions []uint64
	}{
		{
			name:              "get release revisions successfully",
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock google storage operation", t, func() {
				revisions := mockGoogleStorage().GetRevisions()
				assert.Equal(t, tc.expectedRevisions, revisions)
			})
		})
	}
}

func TestGoogleStorage_GetStackBoundRevisions(t *testing.T) {
	testcases := []struct {
		name              string
		stack             string
		expectedRevisions []uint64
	}{
		{
			name:              "get stack bound release revisions successfully",
			stack:             "test_stack",
			expectedRevisions: []uint64{1, 2, 3},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock google storage operation", t, func() {
				revisions := mockGoogleStorage().GetStackBoundRevisions(tc.stack)
				assert.Equal(t, tc.expectedRevisions, revisions)
			})
		})
	}
}

func TestGoogleStorage_GetLatestRevision(t *testing.T) {
	testcases := []struct {
		name             string
		expectedRevision uint64
	}{
		{
			name:             "get latest release revision successfully",
			expectedRevision: 3,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock google storage operation", t, func() {
				revision := mockGoogleStorage().GetLatestRevision()
				assert.Equal(t, tc.expectedRevision, revision)
			})
		})
	}
}

func TestGoogleStorage_Create(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		r       *v1.Release
	}{
		{
			name:    "create release successfully",
			success: true,
			r:       mockRelease(4),
		},
		{
			name:    "failed to create release already exist",
			success: false,
			r:       mockRelease(3),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockGoogleStorageWriteMeta()
				mockGoogleStorageWriteRelease()
				err := mockGoogleStorage().Create(tc.r)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestGoogleStorage_Update(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		r       *v1.Release
	}{
		{
			name:    "update release successfully",
			success: true,
			r:       mockRelease(3),
		},
		{
			name:    "failed to update release not exist",
			success: false,
			r:       mockRelease(4),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockGoogleStorageWriteRelease()
				err := mockGoogleStorage().Update(tc.r)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
