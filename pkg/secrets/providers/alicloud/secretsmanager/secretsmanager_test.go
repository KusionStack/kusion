package secretsmanager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets/providers/alicloud/secretsmanager/fake"
)

func TestGetSecret(t *testing.T) {
	testCases := map[string]struct {
		client      Client
		name        string
		property    string
		expected    []byte
		expectedErr error
	}{
		"GetSecret": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn("t0p-Secret", "text", nil),
			},
			name:        "/beep",
			expected:    []byte(`t0p-Secret`),
			expectedErr: nil,
		},
		"GetSecret_With_Property": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn(`{"bar": "bang"}`, "text", nil),
			},
			name:        "/beep",
			property:    "bar",
			expected:    []byte(`bang`),
			expectedErr: nil,
		},
		"GetSecret_With_NestedProperty": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn(`{"foobar":{"bar":"bang"}}`, "text", nil),
			},
			name:        "/beep",
			property:    "foobar.bar",
			expected:    []byte(`bang`),
			expectedErr: nil,
		},
		"GetSecret_With_Binary": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn([]byte(`t0p-Secret`), "binary", nil),
			},
			name:        "/beep",
			expected:    []byte(`t0p-Secret`),
			expectedErr: nil,
		},
		"GetSecret_With_Property_Binary": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn([]byte(`{"bar":"bang"}`), "binary", nil),
			},
			name:        "/beep",
			property:    "bar",
			expected:    []byte(`bang`),
			expectedErr: nil,
		},
		"GetSecret_With_NestedProperty_Binary": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn([]byte(`{"foobar":{"bar":"bang"}}`), "binary", nil),
			},
			name:        "/beep",
			property:    "foobar.bar",
			expected:    []byte(`bang`),
			expectedErr: nil,
		},
		"GetSecret_With_Error": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn([]byte(`{"foobar":{"bar":"bang"}}`), "binary", errors.New("internal error")),
			},
			name:        "/beep",
			expected:    nil,
			expectedErr: errors.New("internal error"),
		},
		"GetSecret_Property_NotFound": {
			client: &fake.SecretsManagerClient{
				GetSecretInfoFn: fake.NewGetSecretInfoFn([]byte(`{"foobar":{"bar":"bang"}}`), "binary", nil),
			},
			name:        "/beep",
			property:    "foobar.baz",
			expected:    nil,
			expectedErr: fmt.Errorf("key foobar.baz does not exist in secret /beep"),
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
		if diff := cmp.Diff(err, tc.expectedErr, EquateErrors()); diff != "" {
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
		spec        v1.SecretStoreSpec
		expectedErr error
	}{
		"InvalidSecretStoreSpec": {
			spec:        v1.SecretStoreSpec{},
			expectedErr: errors.New(errMissingProviderSpec),
		},
		"InvalidProviderSpec": {
			spec: v1.SecretStoreSpec{
				Provider: &v1.ProviderSpec{},
			},
			expectedErr: errors.New(errMissingAlicloudProvider),
		},
		"ValidVaultProviderSpec": {
			spec: v1.SecretStoreSpec{
				Provider: &v1.ProviderSpec{
					Alicloud: &v1.AlicloudProvider{
						Region: "cn-beijing",
					},
				},
			},
			expectedErr: nil,
		},
	}

	factory := DefaultSecretStoreProvider{}
	for name, tc := range testCases {
		_, err := factory.NewSecretStore(tc.spec)
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
