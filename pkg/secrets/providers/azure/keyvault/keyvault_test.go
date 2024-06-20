package keyvault

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets/providers/azure/keyvault/fake"
)

const (
	jwkPubRSA = `{"kid":"ex","kty":"RSA","key_ops":["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"],"n":"=","e":"="}`
)

var (
	fakeVaultURL = "noop"
	fakeTenantID = "beep"
)

func TestGetSecret(t *testing.T) {
	testCases := map[string]struct {
		client    SecretClient
		name      string
		property  string
		expected  []byte
		expectErr error
	}{
		"GetSecret": {
			client: &fake.SecretClient{
				GetSecretFn: fake.NewGetSecretFn("t0p-Secret"),
			},
			name:      "test-secret",
			expected:  []byte(`t0p-Secret`),
			expectErr: nil,
		},
		"GetSecret_With_Property": {
			client: &fake.SecretClient{
				GetSecretFn: fake.NewGetSecretFn(`{"bar":"bang"}`),
			},
			name:      "test-secret",
			property:  "bar",
			expected:  []byte(`bang`),
			expectErr: nil,
		},
		"GetSecret_Property_NotFound": {
			client: &fake.SecretClient{
				GetSecretFn: fake.NewGetSecretFn(`{"bar":"bang"}`),
			},
			name:      "test-secret",
			property:  "barr",
			expectErr: fmt.Errorf(errPropertyNotExist, "barr", "test-secret"),
		},
		"GetKey": {
			client: &fake.SecretClient{
				GetKeyFn: fake.NewGetKeyFn(jwkPubRSA),
			},
			name:      "key/keyName",
			expected:  []byte(jwkPubRSA),
			expectErr: nil,
		},
		"GetCertificate": {
			client: &fake.SecretClient{
				GetCertificateFn: fake.NewGetCertificateFn("certificate_value"),
			},
			name:      "cert/certName",
			expected:  []byte(`certificate_value`),
			expectErr: nil,
		},
	}

	provider := &v1.AzureKVProvider{
		VaultURL: &fakeVaultURL,
		TenantID: &fakeTenantID,
	}

	for name, tc := range testCases {
		store := &kvSecretStore{
			secretClient: tc.client,
			provider:     provider,
		}
		ref := v1.ExternalSecretRef{
			Name:     tc.name,
			Property: tc.property,
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
		spec        v1.SecretStore
		initEnv     bool
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
			expectedErr: errors.New(errMissingAzureProvider),
		},
		"InvalidAzureKVProviderSpec": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{
					Azure: &v1.AzureKVProvider{
						VaultURL: &fakeVaultURL,
					},
				},
			},
			expectedErr: errors.New(errMissingTenant),
		},
		"NoClientIDSecretEnvFound": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{
					Azure: &v1.AzureKVProvider{
						VaultURL: &fakeVaultURL,
						TenantID: &fakeTenantID,
					},
				},
			},
			expectedErr: errors.New(errMissingClientIDSecret),
		},
		"ValidVaultProviderSpec": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{
					Azure: &v1.AzureKVProvider{
						VaultURL: &fakeVaultURL,
						TenantID: &fakeTenantID,
					},
				},
			},
			initEnv:     true,
			expectedErr: nil,
		},
	}

	factory := DefaultSecretStoreProvider{}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.initEnv {
				cleanup := fake.SetClientIDSecretInEnv()
				defer cleanup()
			}
			_, err := factory.NewSecretStore(&tc.spec)
			if diff := cmp.Diff(err, tc.expectedErr, EquateErrors()); diff != "" {
				t.Errorf("\n%s\ngot unexpected error:\n%s", name, diff)
			}
		})
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
