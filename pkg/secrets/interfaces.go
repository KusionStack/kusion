package secrets

import (
	"context"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// SecretStore provides the interface to interact with various cloud secret manager.
type SecretStore interface {
	// GetSecret retrieves ref secret from various cloud secret manager.
	GetSecret(ctx context.Context, ref v1.ExternalSecretRef) ([]byte, error)
}

// SecretStoreFactory is a factory type for secret store.
type SecretStoreFactory interface {
	// NewSecretStore constructs a usable secret store with specific provider spec.
	NewSecretStore(spec v1.SecretStoreSpec) (SecretStore, error)
}
