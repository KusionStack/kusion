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
func (fss *FakeSecretStore) GetSecret(_ context.Context, _ secrets.ExternalSecretRef) ([]byte, error) {
	return []byte("NOOP"), nil
}

// FakeSecretStoreFactory is the fake implementation of SecretStoreFactory.
type FakeSecretStoreFactory struct{}

// Fake implementation of SecretStoreFactory.NewSecretStore.
func (fsf *FakeSecretStoreFactory) NewSecretStore(_ secrets.SecretStoreSpec) (SecretStore, error) {
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

	fsp := &FakeSecretStoreFactory{}
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
