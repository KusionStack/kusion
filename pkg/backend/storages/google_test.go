package storages

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func TestNewGoogleStorage(t *testing.T) {
	tests := []struct {
		name    string
		config  *v1.BackendGoogleConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &v1.BackendGoogleConfig{
				Credentials: &google.Credentials{
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
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(tt.config.Credentials.JSON))
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewGoogleStorage() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			defer client.Close()

			got, err := NewGoogleStorage(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGoogleStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, got)
				assert.Equal(t, tt.config.Prefix, got.prefix)
			}
		})
	}
}
