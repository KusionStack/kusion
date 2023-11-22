package secrets

import (
	"context"

	secretsapi "kusionstack.io/kusion/pkg/apis/secrets"
)

// SecretStore provides the interface to interact with various cloud secret manager.
type SecretStore interface {
	// GetSecret retrieves ref secret from various cloud secret manager.
	GetSecret(ctx context.Context, ref string) ([]byte, error)
}

// SecretStoreProvider is a factory type for secret store.
type SecretStoreProvider interface {
	// Type returns a string that reflects the type of this provider.
	Type() string
	// NewSecretStore constructs a usable secret store with specific provider spec.
	NewSecretStore(spec *secretsapi.SecretStoreSpec) (SecretStore, error)
}
