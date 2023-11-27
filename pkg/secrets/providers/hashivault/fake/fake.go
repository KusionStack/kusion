package fake

import (
	"context"
	"os"

	vault "github.com/hashicorp/vault/api"
)

type (
	ReadWithDataWithContextFn func(ctx context.Context, path string, data map[string][]string) (*vault.Secret, error)
	Logical                   struct {
		ReadWithDataWithContextFn ReadWithDataWithContextFn
	}
)

func NewReadWithContextFn(secretData map[string]interface{}, err error) ReadWithDataWithContextFn {
	return func(ctx context.Context, path string, data map[string][]string) (*vault.Secret, error) {
		if secretData == nil {
			return nil, err
		}
		secret := &vault.Secret{
			Data: secretData,
		}
		return secret, err
	}
}

func (f Logical) ReadWithDataWithContext(ctx context.Context, path string, data map[string][]string) (*vault.Secret, error) {
	return f.ReadWithDataWithContextFn(ctx, path, data)
}

func SetTokenInEnv() func() {
	oldTokenVal := os.Getenv("VAULT_SERVER_TOKEN")
	os.Setenv("VAULT_SERVER_TOKEN", "fake_token")
	return func() {
		os.Setenv("VAULT_SERVER_TOKEN", oldTokenVal)
	}
}

func SetAlternativeTokenInEnv() func() {
	oldTokenVal := os.Getenv("VAULT_TOKEN")
	os.Setenv("VAULT_TOKEN", "fake_token")
	return func() {
		os.Setenv("VAULT_TOKEN", oldTokenVal)
	}
}
