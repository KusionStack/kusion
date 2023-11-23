package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/secrets"
)

// FakeSecretStore is the fake implementation of SecretStore.
type FakeSecretStore struct{}

// Fake implementation of SecretStore.GetSecret.
func (fss *FakeSecretStore) GetSecret(_ context.Context, _ string) ([]byte, error) {
	return []byte("NOOP"), nil
}

// FakeSecretStoreProvider is the fake implementation of SecretStoreProvider.
type FakeSecretStoreProvider struct{}

// Fake implementation of SecretStoreProvider.Type.
func (fsp *FakeSecretStoreProvider) Type() string {
	return "fake"
}

// Fake implementation of SecretStoreProvider.NewSecretStore.
func (fsp *FakeSecretStoreProvider) NewSecretStore(_ *secrets.SecretStoreSpec) (SecretStore, error) {
	return &FakeSecretStore{}, nil
}

func TestRegister(t *testing.T) {
	testcases := []struct {
		name         string
		providerName string
		shouldPanic  bool
		expExists    bool
		spec         *secrets.ProviderSpec
	}{
		{
			name:        "should panic when given an invalid provider spec",
			shouldPanic: true,
			spec:        &secrets.ProviderSpec{},
		},
		{
			name:         "should register a valid provider",
			providerName: "aws",
			shouldPanic:  false,
			expExists:    true,
			spec: &secrets.ProviderSpec{
				AWS: &secrets.AWSProvider{},
			},
		},
	}

	providers := NewProviders()
	fsp := &FakeSecretStoreProvider{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Register should panic")
					}
				}()
			}

			providers.Register(fsp, tc.spec)
			_, ok := providers.GetProviderByName(tc.providerName)
			assert.Equal(t, tc.expExists, ok, "provider should be registered")
		})
	}
}
