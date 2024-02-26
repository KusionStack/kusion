package secret

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
	// ensure we can get correct secret store provider
	_ "kusionstack.io/kusion/pkg/secrets/providers/register"
)

var testProject = "helloworld"

func initGeneratorRequest(
	project string,
	secrets map[string]workload.Secret,
	secretStoreSpec *apiv1.SecretStoreSpec,
) *GeneratorRequest {
	return &GeneratorRequest{
		Project: project,
		Workload: &workload.Workload{
			Service: &workload.Service{
				Base: workload.Base{
					Secrets: secrets,
				},
			},
		},
		Namespace:       project,
		SecretStoreSpec: secretStoreSpec,
	}
}

func initSecretStoreSpec(data []apiv1.FakeProviderData) *apiv1.SecretStoreSpec {
	return &apiv1.SecretStoreSpec{
		Provider: &apiv1.ProviderSpec{
			Fake: &apiv1.FakeProvider{
				Data: data,
			},
		},
	}
}

func TestGenerateSecret(t *testing.T) {
	tests := map[string]struct {
		secretName string
		secretType string
		secretData map[string]string

		expectErr string
	}{
		"create_basic_auth_secret": {
			secretName: "secret-basic-auth",
			secretType: "basic",
			secretData: map[string]string{
				"username": "admin",
				"password": "t0p-Secret",
			},
		},
		"create_basic_auth_secret_empty_input": {
			secretName: "secret-basic-auth",
			secretType: "basic",
			secretData: map[string]string{},
		},
		"create_token_secret": {
			secretName: "secret-token",
			secretType: "token",
			secretData: map[string]string{
				"token": "YmFyCg==",
			},
		},
		"create_token_secret_empty_input": {
			secretName: "secret-token",
			secretType: "token",
			secretData: map[string]string{},
		},
		"create_opaque_secret": {
			secretName: "empty-secret",
			secretType: "opaque",
			secretData: map[string]string{},
		},
		"create_opaque_secret_any_info": {
			secretName: "empty-secret",
			secretType: "opaque",
			secretData: map[string]string{
				"accessKey": "dHJ1ZQ==",
			},
		},
		"create_certificate_secret": {
			secretName: "secret-tls",
			secretType: "certificate",
			secretData: map[string]string{
				"tls.crt": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNVakNDQWJz",
				"tls.key": "RXhhbXBsZSBkYXRhIGZvciB0aGUgVExTIGNydCBmaWVsZA==",
			},
		},
		"create_invalid_secret_invalid_type": {
			secretName: "invalid-tls",
			secretType: "cred",
			expectErr:  "unrecognized secret type cred",
		},
	}

	// run all the tests
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			secrets := map[string]workload.Secret{
				name: {
					Type: test.secretType,
					Data: test.secretData,
				},
			}
			context := initGeneratorRequest(testProject, secrets, nil)
			generator, _ := NewSecretGenerator(context)
			err := generator.Generate(&apiv1.Intent{})
			if test.expectErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectErr)
			}
		})
	}
}

func TestGenerateSecretWithExternalRef(t *testing.T) {
	tests := map[string]struct {
		secretName string
		secretType string
		secretData map[string]string

		providerData []apiv1.FakeProviderData

		expectErr string
	}{
		"create_external_secret": {
			secretName: "api-auth",
			secretType: "external",
			secretData: map[string]string{
				"accessKey": "ref://api-auth-info/accessKey?version=1",
				"secretKey": "ref://api-auth-info/secretKey?version=1",
			},
			providerData: []apiv1.FakeProviderData{
				{
					Key:     "api-auth-info",
					Value:   `{"accessKey":"some sensitive info","secretKey":"*******"}`,
					Version: "1",
				},
			},
		},
		"create_external_secret_not_found": {
			secretName: "access-token",
			secretType: "external",
			secretData: map[string]string{
				"accessToken": "ref://token?version=1",
			},
			providerData: []apiv1.FakeProviderData{
				{
					Key:     "token-info",
					Value:   "some sensitive info",
					Version: "1",
				},
			},
			expectErr: "Secret does not exist",
		},
	}

	// run all the tests
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			secrets := map[string]workload.Secret{
				name: {
					Type: test.secretType,
					Data: test.secretData,
				},
			}
			secretStoreSpec := initSecretStoreSpec(test.providerData)
			context := initGeneratorRequest(testProject, secrets, secretStoreSpec)
			generator, _ := NewSecretGenerator(context)
			err := generator.Generate(&apiv1.Intent{})
			if test.expectErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectErr)
			}
		})
	}
}

func TestParseExternalSecretDataRef(t *testing.T) {
	tests := []struct {
		name       string
		dataRefStr string
		want       *apiv1.ExternalSecretRef
		wantErr    bool
	}{
		{
			name:       "invalid data ref string",
			dataRefStr: "$%#//invalid",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "only secret name",
			dataRefStr: "ref://secret-name",
			want: &apiv1.ExternalSecretRef{
				Name: "secret-name",
			},
			wantErr: false,
		},
		{
			name:       "secret name with version",
			dataRefStr: "ref://secret-name?version=1",
			want: &apiv1.ExternalSecretRef{
				Name:    "secret-name",
				Version: "1",
			},
			wantErr: false,
		},
		{
			name:       "secret name with property and version",
			dataRefStr: "ref://secret-name/property?version=1",
			want: &apiv1.ExternalSecretRef{
				Name:     "secret-name",
				Property: "property",
				Version:  "1",
			},
			wantErr: false,
		},
		{
			name:       "nested secret name with property and version",
			dataRefStr: "ref://customer/acme/customer_name?version=1",
			want: &apiv1.ExternalSecretRef{
				Name:     "customer/acme",
				Property: "customer_name",
				Version:  "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExternalSecretDataRef(tt.dataRefStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseExternalSecretDataRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseExternalSecretDataRef() got = %v, want %v", got, tt.want)
			}
		})
	}
}
