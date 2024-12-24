package storages

import (
	"context"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"

	googlestorage "cloud.google.com/go/storage"
	googleauth "golang.org/x/oauth2/google"
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
		meta:   mockWorkspacesMetaData(),
	}
}

func mockGoogleStorageWriteMeta() {
	mockey.Mock((*GoogleStorage).writeMeta).Return(nil).Build()
}

func mockGoogleStorageWriteWorkspace() {
	mockey.Mock((*GoogleStorage).writeWorkspace).Return(nil).Build()
}

func TestGoogleStorage_Create(t *testing.T) {
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
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockGoogleStorageWriteMeta()
				mockGoogleStorageWriteWorkspace()
				err := mockGoogleStorage().Create(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestGoogleStorage_Update(t *testing.T) {
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
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockGoogleStorageWriteWorkspace()
				err := mockGoogleStorage().Update(tc.workspace)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestGoogleStorage_Delete(t *testing.T) {
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
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockey.Mock((*googlestorage.ObjectHandle).Delete).Return(nil).Build()
				mockGoogleStorageWriteMeta()
				err := mockGoogleStorage().Delete(tc.wsName)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}

func TestGoogleStorage_GetNames(t *testing.T) {
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
			mockey.PatchConvey("mock google storage operation", t, func() {
				wsNames, err := mockGoogleStorage().GetNames()
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.Equal(t, tc.expectedNames, wsNames)
				}
			})
		})
	}
}

func TestGoogleStorage_GetCurrent(t *testing.T) {
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
			mockey.PatchConvey("mock google storage operation", t, func() {
				current, err := mockGoogleStorage().GetCurrent()
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.Equal(t, tc.expectedCurrent, current)
				}
			})
		})
	}
}

func TestGoogleStorage_SetCurrent(t *testing.T) {
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
			mockey.PatchConvey("mock google storage operation", t, func() {
				mockGoogleStorageWriteMeta()
				err := mockGoogleStorage().SetCurrent(tc.current)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
