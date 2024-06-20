package fake

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets"
)

func TestGetSecret(t *testing.T) {
	p := &DefaultSecretStoreProvider{}
	testCases := []struct {
		name     string
		input    []v1.FakeProviderData
		ref      v1.ExternalSecretRef
		expErr   string
		expValue string
		wantErr  bool
	}{
		{
			name:  "return err when not found",
			input: []v1.FakeProviderData{},
			ref: v1.ExternalSecretRef{
				Name:    "/secret-name",
				Version: "v2",
			},
			expErr: secrets.NoSecretErr.Error(),
		},
		{
			name: "get correct value from multiple versions",
			input: []v1.FakeProviderData{
				{
					Key:     "/beep",
					Value:   "one",
					Version: "v1",
				},
				{
					Key:   "/bar",
					Value: "xxxxx",
				},
				{
					Key:     "/beep",
					Value:   "two",
					Version: "v2",
				},
			},
			ref: v1.ExternalSecretRef{
				Name:    "/beep",
				Version: "v2",
			},
			expValue: "two",
		},
		{
			name: "get correct value from multiple properties",
			input: []v1.FakeProviderData{
				{
					Key:   "junk",
					Value: "xxxxx",
				},
				{
					Key:   "/customer",
					Value: `{"name":"Tony","age":"24"}`,
				},
			},
			ref: v1.ExternalSecretRef{
				Name:     "/customer",
				Property: "name",
			},
			expValue: "Tony",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ss, _ := p.NewSecretStore(&v1.SecretStore{
				Provider: &v1.ProviderSpec{
					Fake: &v1.FakeProvider{
						Data: tt.input,
					},
				},
			})
			got, err := ss.GetSecret(context.Background(), tt.ref)
			if len(tt.expErr) > 0 && tt.expErr != err.Error() {
				t.Errorf("expected error %s, got %s", tt.expErr, err.Error())
			} else if len(tt.expErr) == 0 && err != nil {
				t.Errorf("unexpected error %v", err)
			}
			if len(tt.expValue) > 0 && tt.expValue != string(got) {
				t.Errorf("expected result %s, got %s", tt.expValue, string(got))
			}
		})
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
			expectedErr: errors.New(errMissingFakeProvider),
		},
		"ValidFakeProviderSpec": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{
					Fake: &v1.FakeProvider{},
				},
			},
			expectedErr: nil,
		},
		"ValidFakeProviderSpec_WithData": {
			spec: v1.SecretStore{
				Provider: &v1.ProviderSpec{
					Fake: &v1.FakeProvider{
						Data: []v1.FakeProviderData{
							{
								Key:     "secret-name",
								Value:   "some sensitive info",
								Version: "1",
							},
						},
					},
				},
			},
			expectedErr: nil,
		},
	}

	provider := DefaultSecretStoreProvider{}
	for name, tc := range testCases {
		_, err := provider.NewSecretStore(&tc.spec)
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
