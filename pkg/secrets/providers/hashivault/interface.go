package hashivault

import (
	"context"

	vault "github.com/hashicorp/vault/api"
)

// Logical is a testable interface for performing logical backend operations on Vault.
type Logical interface {
	ReadWithDataWithContext(ctx context.Context, path string, data map[string][]string) (*vault.Secret, error)
}
