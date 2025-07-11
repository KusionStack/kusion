package secrets

import (
	"context"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// SecretStore provides the interface to interact with various cloud secret manager.
type SecretStore interface {
	// GetSecret retrieves ref secret from various cloud secret manager.
	GetSecret(ctx context.Context, ref v1.ExternalSecretRef) ([]byte, error)
	// SetSecret sets ref secret to various cloud secret manager.
	SetSecret(ctx context.Context, ref v1.ExternalSecretRef, data []byte) error
}

// SecretStoreProvider is a factory type for secret store.
type SecretStoreProvider interface {
	// NewSecretStore constructs a usable secret store with specific provider spec.
	NewSecretStore(spec *v1.SecretStore) (SecretStore, error)
}

var NoSecretErr = NoSecretError{}

// NoSecretError will be returned when GetSecret call can not find the
// desired secret.
type NoSecretError struct{}

func (NoSecretError) Error() string {
	return "Secret does not exist"
}
