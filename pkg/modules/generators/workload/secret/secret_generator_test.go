package secret

import (
	"testing"

	"github.com/stretchr/testify/require"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	// ensure we can get correct secret store provider
	_ "kusionstack.io/kusion/pkg/secrets/providers/register"
)

var testProject = "helloworld"

func initGeneratorRequest(
	project string,
	secrets map[string]v1.Secret,
	secretStoreSpec *v1.SecretStore,
) *GeneratorRequest {
	return &GeneratorRequest{
		Project: project,
		Workload: &v1.Workload{
			Service: &v1.Service{
				Base: v1.Base{
					Secrets: secrets,
				},
			},
		},
		Namespace:   project,
		SecretStore: secretStoreSpec,
	}
}

func initSecretStoreSpec(data []v1.FakeProviderData) *v1.SecretStore {
	return &v1.SecretStore{
		Provider: &v1.ProviderSpec{
			Fake: &v1.FakeProvider{
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
			secrets := map[string]v1.Secret{
				name: {
					Type: test.secretType,
					Data: test.secretData,
				},
			}
			context := initGeneratorRequest(testProject, secrets, nil)
			generator, _ := NewSecretGenerator(context)
			err := generator.Generate(&v1.Spec{})
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

		providerData []v1.FakeProviderData

		expectErr string
	}{
		"create_external_secret": {
			secretName: "api-auth",
			secretType: "external",
			secretData: map[string]string{
				"accessKey": "ref://api-auth-info/accessKey?version=1",
				"secretKey": "ref://api-auth-info/secretKey?version=1",
			},
			providerData: []v1.FakeProviderData{
				{
					Key:     "api-auth-info",
					Value:   `{"accessKey":"some sensitive info","secretKey":"*******"}`,
					Version: "1",
				},
			},
		},
	}

	// run all the tests
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			secrets := map[string]v1.Secret{
				name: {
					Type: test.secretType,
					Data: test.secretData,
				},
			}
			secretStoreSpec := initSecretStoreSpec(test.providerData)
			context := initGeneratorRequest(testProject, secrets, secretStoreSpec)
			generator, _ := NewSecretGenerator(context)
			err := generator.Generate(&v1.Spec{})
			if test.expectErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectErr)
			}
		})
	}
}
