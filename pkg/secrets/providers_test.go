package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// FakeSecretStore is the fake implementation of SecretStore.
type FakeSecretStore struct{}

// Fake implementation of SecretStore.GetSecret.
func (fss *FakeSecretStore) GetSecret(_ context.Context, _ v1.ExternalSecretRef) ([]byte, error) {
	return []byte("NOOP"), nil
}

// FakeSecretStoreProvider is the fake implementation of SecretStoreProvider.
type FakeSecretStoreProvider struct{}

// Fake implementation of SecretStoreProvider.NewSecretStore.
func (fsf *FakeSecretStoreProvider) NewSecretStore(_ v1.SecretStoreSpec) (SecretStore, error) {
	return &FakeSecretStore{}, nil
}

func TestRegister(t *testing.T) {
	testcases := []struct {
		name         string
		providerName string
		shouldPanic  bool
		expExists    bool
		spec         *v1.ProviderSpec
	}{
		{
			name:        "should panic when given an invalid provider spec",
			shouldPanic: true,
			spec:        &v1.ProviderSpec{},
		},
		{
			name:         "should register a valid provider",
			providerName: "aws",
			shouldPanic:  false,
			expExists:    true,
			spec: &v1.ProviderSpec{
				AWS: &v1.AWSProvider{},
			},
		},
	}

	fsp := &FakeSecretStoreProvider{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("register should panic")
					}
				}()
			}

			Register(fsp, tc.spec)
			_, ok := GetProviderByName(tc.providerName)
			assert.Equal(t, tc.expExists, ok, "provider should be registered")
		})
	}
}
