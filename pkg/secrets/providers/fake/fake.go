package fake

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets"
)

const (
	errMissingProviderSpec  = "secret store spec is missing provider"
	errMissingFakeProvider  = "invalid provider spec. Missing Fake field in secret store provider spec"
	errMethodNotImplemented = "method not implemented. secret provider: %s, method: %s"
)

type SecretData struct {
	Value    string
	Version  string
	ValueMap map[string]string
}

// DefaultSecretStoreProvider should implement the secrets.SecretStoreProvider interface
var _ secrets.SecretStoreProvider = &DefaultSecretStoreProvider{}

// smSecretStore should implement the secrets.SecretStore interface
var _ secrets.SecretStore = &fakeSecretStore{}

type DefaultSecretStoreProvider struct{}

// NewSecretStore constructs a fake secret store instance.
func (p *DefaultSecretStoreProvider) NewSecretStore(spec *v1.SecretStore) (secrets.SecretStore, error) {
	providerSpec := spec.Provider
	if providerSpec == nil {
		return nil, fmt.Errorf(errMissingProviderSpec)
	}
	if providerSpec.Fake == nil {
		return nil, fmt.Errorf(errMissingFakeProvider)
	}

	dataMap := make(map[string]*SecretData)
	for _, data := range providerSpec.Fake.Data {
		key := mapKey(data.Key, data.Version)
		dataMap[key] = &SecretData{
			Value:   data.Value,
			Version: data.Version,
		}
		if data.ValueMap != nil {
			dataMap[key].ValueMap = data.ValueMap
		}
	}

	return &fakeSecretStore{dataMap: dataMap}, nil
}

type fakeSecretStore struct {
	dataMap map[string]*SecretData
}

// GetSecret retrieves ref secret value from backend data map.
func (f *fakeSecretStore) GetSecret(_ context.Context, ref v1.ExternalSecretRef) ([]byte, error) {
	data, ok := f.dataMap[mapKey(ref.Name, ref.Version)]
	if !ok || data.Version != ref.Version {
		return nil, secrets.NoSecretErr
	}

	if ref.Property != "" {
		val := gjson.Get(data.Value, ref.Property)
		if !val.Exists() {
			return nil, secrets.NoSecretErr
		}

		return []byte(val.String()), nil
	}

	return []byte(data.Value), nil
}

// SetSecret sets ref secret value to backend data map.
func (f *fakeSecretStore) SetSecret(ctx context.Context, ref v1.ExternalSecretRef, secretValue []byte) error {
	return fmt.Errorf(errMethodNotImplemented, "fake", "SetSecret")
}

func mapKey(key, version string) string {
	// Add the version suffix to preserve entries with the old versions as well.
	return fmt.Sprintf("%v%v", key, version)
}

func init() {
	secrets.Register(&DefaultSecretStoreProvider{}, &v1.ProviderSpec{
		Fake: &v1.FakeProvider{},
	})
}
