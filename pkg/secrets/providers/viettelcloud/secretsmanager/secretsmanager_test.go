package secretsmanager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets/providers/viettelcloud/secretsmanager/fake"
)

func TestGetSecret(t *testing.T) {
	testCases := map[string]struct {
		client    Client
		name      string
		property  string
		expected  []byte
		expectErr error
	}{
		"GetSecret_With_Property": {
			client: &fake.SecretsManagerClient{
				SecretManagerSecretsRetrieveWithResponseFn: fake.NewSecretManagerSecretsRetrieveWithResponseFn(
					map[string]interface{}{
						"password": "p455w0rd",
					}, "key-value", nil),
			},
			name:      "beep",
			property:  "password",
			expected:  []byte(`p455w0rd`),
			expectErr: nil,
		},
		"GetSecret_Property_NotFound": {
			client: &fake.SecretsManagerClient{
				SecretManagerSecretsRetrieveWithResponseFn: fake.NewSecretManagerSecretsRetrieveWithResponseFn(
					map[string]interface{}{
						"password": "p455w0rd",
					}, "key-value", nil),
			},
			name:      "beep",
			property:  "notfound",
			expected:  []byte(``),
			expectErr: fmt.Errorf("key notfound does not exist in secret beep"),
		},
	}

	for name, tc := range testCases {
		store := &smSecretStore{client: tc.client}
		ref := v1.ExternalSecretRef{
			Name:     tc.name,
			Property: tc.property,
		}
		if name != "GetSecret_Property_NotFound" {
			continue
		}
		actual, err := store.GetSecret(context.TODO(), ref)
		t.Logf("actual: %v, err: %v", actual, err)
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
		spec        v1.SecretStore
		expectedErr error
	}{
		"InvalidSecretStoreSpec": {
			spec:        v1.SecretStore{},
			expectedErr: errors.New(errMissingProviderSpec),
		},
		"InvalidProviderSpec": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{},
			},
			expectedErr: errors.New(errMissingViettelCloudProvider),
		},
		"ValidProviderSpec": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{
					ViettelCloud: &v1.ViettelCloudProvider{
						CmpURL:    "https://console.viettelcloud.vn/api",
						ProjectID: "00000000-0000-0000-0000-000000000000",
					},
				},
			},
			expectedErr: nil,
		},
	}
	factory := DefaultSecretStoreProvider{}
	for name, tc := range testCases {
		_, err := factory.NewSecretStore(&tc.spec)
		if diff := cmp.Diff(err, tc.expectedErr, EquateErrors()); diff != "" {
			t.Errorf("\n%s\ngot unexpected error: \n%s", name, diff)
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
