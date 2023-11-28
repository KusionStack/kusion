package secret

import (
	"testing"

	"github.com/stretchr/testify/require"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

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

	project := &project.Project{
		ProjectConfiguration: project.ProjectConfiguration{
			Name: "helloworld",
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
			intent := &intent.Intent{}
			generator, _ := NewSecretGenerator(project, secrets)
			err := generator.Generate(intent)
			if test.expectErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectErr)
			}
		})
	}
}
