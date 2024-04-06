package hashivault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets/providers/hashivault/fake"
)

var mountPath = "secret"

func makeValidVaultSecretStore(v v1.VaultKVStoreVersion) *vaultSecretStore {
	return &vaultSecretStore{
		provider: &v1.VaultProvider{
			Path:    &mountPath,
			Version: v,
		},
	}
}

func makeExternalSecretRef(path, property, version string) v1.ExternalSecretRef {
	return v1.ExternalSecretRef{
		Name:     path,
		Property: property,
		Version:  version,
	}
}

func TestGetSecret(t *testing.T) {
	secretData := map[string]interface{}{
		"username": "admin",
		"password": "t0p-Secret",
	}
	secretDataWithNilValue := map[string]interface{}{
		"username": "admin",
		"password": "t0p-Secret",
		"token":    nil,
	}
	secretDataWithNestedValue := map[string]interface{}{
		"username":      "admin",
		"password":      "t0p-Secret",
		"aws.accessKey": "access_key",
		"gcp": map[string]string{
			"accessKey": "foo",
			"secretKey": "bar",
		},
		"list_of_values": []string{
			"one",
			"two",
			"three",
		},
		"json_number": json.Number("72"),
	}
	secretNestedData := map[string]interface{}{
		"data": map[string]interface{}{
			"username": "admin",
			"password": "t0p-Secret",
		},
	}
	testCases := map[string]struct {
		provider  *v1.VaultProvider
		logical   Logical
		path      string
		property  string
		version   string
		expected  []byte
		expectErr error
	}{
		"V1_ReadSecret": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretData, nil),
			},
			path:      "secret/path",
			property:  "password",
			expected:  []byte(`t0p-Secret`),
			expectErr: nil,
		},
		"V1_ReadSecret_NoProperty": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretData, nil),
			},
			path:      "secret/path",
			expected:  []byte(`{"password":"t0p-Secret","username":"admin"}`),
			expectErr: nil,
		},
		"V1_ReadSecret_NoProperty_NilValue": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretDataWithNilValue, nil),
			},
			path:      "secret/path",
			expected:  []byte(`{"password":"t0p-Secret","token":null,"username":"admin"}`),
			expectErr: nil,
		},
		"V1_ReadSecret_NestedValue": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretDataWithNestedValue, nil),
			},
			path:      "secret/path",
			property:  "aws.accessKey",
			expected:  []byte(`access_key`),
			expectErr: nil,
		},
		"V1_ReadSecret_NestedValue_NestedProperty": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretDataWithNestedValue, nil),
			},
			path:      "secret/path",
			property:  "gcp.accessKey",
			expected:  []byte(`foo`),
			expectErr: nil,
		},
		"V1_ReadSecret_SliceValue": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretDataWithNestedValue, nil),
			},
			path:      "secret/path",
			property:  "list_of_values",
			expected:  []byte("one\ntwo\nthree"),
			expectErr: nil,
		},
		"V1_ReadSecret_JsonNumber": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretDataWithNestedValue, nil),
			},
			path:      "secret/path",
			property:  "json_number",
			expected:  []byte("72"),
			expectErr: nil,
		},
		"V1_ReadSecret_NilData": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV1).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(nil, nil),
			},
			path:      "secret/path",
			expected:  []byte("{}"),
			expectErr: nil,
		},
		"V2_ReadSecret": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV2).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretNestedData, nil),
			},
			path:      "secret/path",
			property:  "username",
			expected:  []byte("admin"),
			expectErr: nil,
		},
		"V2_ReadSecret_WithVersion": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV2).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretNestedData, nil),
			},
			path:      "secret/path",
			property:  "username",
			version:   "1",
			expected:  []byte("admin"),
			expectErr: nil,
		},
		"V2_ReadSecret_ErrFormat": {
			provider: makeValidVaultSecretStore(v1.VaultKVStoreV2).provider,
			logical: &fake.Logical{
				ReadWithDataWithContextFn: fake.NewReadWithContextFn(secretData, nil),
			},
			path:      "secret/path",
			expectErr: errors.New(errParseDataField),
		},
	}

	for name, tc := range testCases {
		store := &vaultSecretStore{
			provider: tc.provider,
			logical:  tc.logical,
		}
		ref := makeExternalSecretRef(tc.path, tc.property, tc.version)
		actual, err := store.GetSecret(context.Background(), ref)
		if diff := cmp.Diff(err, tc.expectErr, EquateErrors()); diff != "" {
			t.Errorf("\n%s\ngot unexpected error:\n%s", name, diff)
		}
		if diff := cmp.Diff(string(actual), string(tc.expected)); diff != "" {
			fmt.Println(diff)
			t.Errorf("\n%s\nget unexpected data: \n%s", name, diff)
		}
	}
}

func TestBuildPath(t *testing.T) {
	storeV2 := makeValidVaultSecretStore(v1.VaultKVStoreV2)
	otherMountPath := "secret/path"
	storeV2.provider.Path = &otherMountPath
	storeV2NoPath := makeValidVaultSecretStore(v1.VaultKVStoreV2)
	storeV2NoPath.provider.Path = nil

	storeV1 := makeValidVaultSecretStore(v1.VaultKVStoreV1)
	storeV1.provider.Path = &otherMountPath
	storeV1NoPath := makeValidVaultSecretStore(v1.VaultKVStoreV1)
	storeV1NoPath.provider.Path = nil

	testCases := map[string]struct {
		store    *vaultSecretStore
		path     string
		expected string
	}{
		"V2_NoMountPath_NoData": {
			store:    storeV2NoPath,
			path:     "secret/test/path",
			expected: "secret/data/test/path",
		},
		"V2_NoMountPath_WithData": {
			store:    storeV2NoPath,
			path:     "secret/path/data/test/first",
			expected: "secret/path/data/test/first",
		},
		"V2_NoData": {
			store:    storeV2,
			path:     "secret/path/test",
			expected: "secret/path/data/test",
		},
		"V2_WithMountPath_WithData": {
			store:    storeV2,
			path:     "secret/path/data/test",
			expected: "secret/path/data/test",
		},
		"V2_Simple_Key_Path": {
			store:    storeV2,
			path:     "test",
			expected: "secret/path/data/test",
		},
		"V1_NoMountPath": {
			store:    storeV1NoPath,
			path:     "secret/test",
			expected: "secret/test",
		},
		"V1_WithMountPath": {
			store:    storeV1,
			path:     "secret/path/test",
			expected: "secret/path/test",
		},
	}

	for name, tc := range testCases {
		actual := tc.store.buildPath(tc.path)
		if actual != tc.expected {
			t.Errorf("%s mismatch, expected %s, actual got %s", name, tc.expected, actual)
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
			expectedErr: errors.New(errInvalidVaultSecretStore),
		},
		"InvalidProviderSpec": {
			spec: v1.SecretStoreSpec{
				Provider: &v1.ProviderSpec{},
			},
			expectedErr: errors.New(errInvalidVaultSecretStore),
		},
		"ValidVaultProviderSpec": {
			spec: v1.SecretStoreSpec{
				Provider: &v1.ProviderSpec{
					Vault: &v1.VaultProvider{
						Server: "https://127.0.0.1:8200",
					},
				},
			},
			expectedErr: nil,
		},
		"ValidVaultProviderSpec_WithToken": {
			spec: v1.SecretStoreSpec{
				Provider: &v1.ProviderSpec{
					Vault: &v1.VaultProvider{
						Server: "https://127.0.0.1:8200",
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
			t.Errorf("\n%s\ngot unexpected error:\n%s", name, diff)
		}
	}
}

func TestGetVaultToken(t *testing.T) {
	t.Run("Test Current Token Env Var", func(t *testing.T) {
		cleanup := fake.SetTokenInEnv()
		defer cleanup()

		vaultToken := getVaultToken()
		if vaultToken != "fake_token" {
			t.Errorf("export 'faketoken': got %q", vaultToken)
		}
	})

	t.Run("Test Alternative Token Env Var", func(t *testing.T) {
		cleanup := fake.SetAlternativeTokenInEnv()
		defer cleanup()

		vaultToken := getVaultToken()
		if vaultToken != "fake_token" {
			t.Errorf("export 'faketoken': got %q", vaultToken)
		}
	})
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
