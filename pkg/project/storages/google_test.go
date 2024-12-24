package storages

import (
	"context"
	"testing"

	googlestorage "cloud.google.com/go/storage"
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

func TestGoogleStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		prefix     string
		objects    []string
		want       map[string][]string
		wantErr    bool
	}{
		{
			name:       "error listing objects",
			bucketName: "error-bucket",
			prefix:     "error-prefix",
			objects:    nil,
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := mockGoogleBucketHandle()
			gs := &GoogleStorage{
				bucket: *bucket,
				prefix: tt.prefix,
			}
			got, err := gs.Get()
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleStorage.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
