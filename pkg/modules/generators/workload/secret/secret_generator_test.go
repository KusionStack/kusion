package secret

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
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

	project := &apiv1.Project{
		Name: "helloworld",
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
			intent := &apiv1.Intent{}
			context := modules.GeneratorContext{
				Project: project,
				Application: &inputs.AppConfiguration{
					Workload: &workload.Workload{
						Service: &workload.Service{
							Base: workload.Base{
								Secrets: secrets,
							},
						},
					},
				},
				Namespace: project.Name,
			}
			generator, _ := NewSecretGenerator(context)
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
