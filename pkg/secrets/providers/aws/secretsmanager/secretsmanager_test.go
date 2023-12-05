package secretsmanager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	secretsapi "kusionstack.io/kusion/pkg/apis/secrets"
	"kusionstack.io/kusion/pkg/secrets/providers/aws/secretsmanager/fake"
)

func TestGetSecret(t *testing.T) {
	testCases := map[string]struct {
		client    Client
		name      string
		property  string
		version   string
		expected  []byte
		expectErr error
	}{
		"GetSecret": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn("t0p-Secret", "string", nil),
			},
			name:      "/beep",
			expected:  []byte(`t0p-Secret`),
			expectErr: nil,
		},
		"GetSecret_With_Version": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn("t0p-Secret", "string", nil),
			},
			name:      "/beep",
			version:   "v1",
			expected:  []byte(`t0p-Secret`),
			expectErr: nil,
		},
		"GetSecret_With_UUID_Version": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn("t0p-Secret", "string", nil),
			},
			name:      "/beep",
			version:   "uuid/xyz",
			expected:  []byte(`t0p-Secret`),
			expectErr: nil,
		},
		"GetSecret_With_Property": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn(`{"bar": "bang"}`, "string", nil),
			},
			name:      "/beep",
			property:  "bar",
			expected:  []byte(`bang`),
			expectErr: nil,
		},
		"GetSecret_With_NestedProperty": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn(`{"foobar":{"bar":"bang"}}`, "string", nil),
			},
			name:      "/beep",
			property:  "foobar.bar",
			expected:  []byte(`bang`),
			expectErr: nil,
		},
		"GetSecret_With_Binary": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn([]byte(`t0p-Secret`), "binary", nil),
			},
			name:      "/beep",
			expected:  []byte(`t0p-Secret`),
			expectErr: nil,
		},
		"GetSecret_With_Property_Binary": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn([]byte(`{"bar":"bang"}`), "binary", nil),
			},
			name:      "/beep",
			property:  "bar",
			expected:  []byte(`bang`),
			expectErr: nil,
		},
		"GetSecret_With_NestedProperty_Binary": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn([]byte(`{"foobar":{"bar":"bang"}}`), "binary", nil),
			},
			name:      "/beep",
			property:  "foobar.bar",
			expected:  []byte(`bang`),
			expectErr: nil,
		},
		"GetSecret_With_Error": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn([]byte(`{"foobar":{"bar":"bang"}}`), "binary", errors.New("internal error")),
			},
			name:      "/beep",
			expected:  nil,
			expectErr: errors.New("internal error"),
		},
		"GetSecret_Property_NotFound": {
			client: &fake.SecretsManagerClient{
				GetSecretValueFn: fake.NewGetSecretValueFn([]byte(`{"foobar":{"bar":"bang"}}`), "binary", nil),
			},
			name:      "/beep",
			property:  "foobar.baz",
			expected:  nil,
			expectErr: fmt.Errorf("key foobar.baz does not exist in secret /beep"),
		},
	}

	for name, tc := range testCases {
		store := &smSecretStore{client: tc.client}
		ref := secretsapi.ExternalSecretRef{
			Name:     tc.name,
			Version:  tc.version,
			Property: tc.property,
		}
		if name != "GetSecret_Property_NotFound" {
			continue
		}
		actual, err := store.GetSecret(context.TODO(), ref)
		if diff := cmp.Diff(err, tc.expectErr, EquateErrors()); diff != "" {
			t.Errorf("\n%s\ngot unexpected error:\n%s", name, diff)
		}
		if diff := cmp.Diff(string(actual), string(tc.expected)); diff != "" {
			fmt.Println(diff)
			t.Errorf("\n%s\nget unexpected data: \n%s", name, diff)
		}
	}
}

func TestNewSecretStore(t *testing.T) {
	testCases := map[string]struct {
		spec        secretsapi.SecretStoreSpec
		expectedErr error
	}{
		"InvalidSecretStoreSpec": {
			spec:        secretsapi.SecretStoreSpec{},
			expectedErr: errors.New(errMissingProviderSpec),
		},
		"InvalidProviderSpec": {
			spec: secretsapi.SecretStoreSpec{
				Provider: &secretsapi.ProviderSpec{},
			},
			expectedErr: errors.New(errMissingAWSProvider),
		},
		"ValidVaultProviderSpec": {
			spec: secretsapi.SecretStoreSpec{
				Provider: &secretsapi.ProviderSpec{
					AWS: &secretsapi.AWSProvider{
						Region: "us-east-1",
					},
				},
			},
			expectedErr: nil,
		},
	}

	factory := DefaultFactory{}
	for name, tc := range testCases {
		_, err := factory.NewSecretStore(tc.spec)
		if diff := cmp.Diff(err, tc.expectedErr, EquateErrors()); diff != "" {
			t.Errorf("\n%s\ngot unexpected error:\n%s", name, diff)
		}
	}
}

// EquateErrors returns true if the supplied errors are of the same type and
// produce same error message.
func EquateErrors() cmp.Option {
	return cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == nil && b == nil
		}

		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		if av.Type() != bv.Type() {
			return false
		}

		return a.Error() == b.Error()
	})
}
