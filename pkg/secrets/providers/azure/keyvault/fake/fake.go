package fake

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
)

type (
	GetSecretFn      func(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (keyvault.SecretBundle, error)
	GetKeyFn         func(ctx context.Context, vaultBaseURL string, keyName string, keyVersion string) (keyvault.KeyBundle, error)
	GetCertificateFn func(ctx context.Context, vaultBaseURL string, certificateName string, certificateVersion string) (keyvault.CertificateBundle, error)
	SecretClient     struct {
		GetSecretFn      GetSecretFn
		GetKeyFn         GetKeyFn
		GetCertificateFn GetCertificateFn
	}
)

func NewGetSecretFn(secretString string) GetSecretFn {
	return func(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (keyvault.SecretBundle, error) {
		return keyvault.SecretBundle{
			Value: &secretString,
		}, nil
	}
}

func NewGetKeyFn(key string) GetKeyFn {
	return func(ctx context.Context, vaultBaseURL string, keyName string, keyVersion string) (keyvault.KeyBundle, error) {
		return keyvault.KeyBundle{
			Key: newJSONWebKey([]byte(key)),
		}, nil
	}
}

func NewGetCertificateFn(certificate string) GetCertificateFn {
	return func(ctx context.Context, vaultBaseURL string, certificateName string, certificateVersion string) (keyvault.CertificateBundle, error) {
		byteStr := []byte(certificate)
		return keyvault.CertificateBundle{
			Cer: &byteStr,
		}, nil
	}
}

func (sc *SecretClient) GetSecret(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (keyvault.SecretBundle, error) {
	return sc.GetSecretFn(ctx, vaultBaseURL, secretName, secretVersion)
}

func (sc *SecretClient) GetKey(ctx context.Context, vaultBaseURL string, keyName string, keyVersion string) (keyvault.KeyBundle, error) {
	return sc.GetKeyFn(ctx, vaultBaseURL, keyName, keyVersion)
}

func (sc *SecretClient) GetCertificate(ctx context.Context, vaultBaseURL string, certificateName string, certificateVersion string) (keyvault.CertificateBundle, error) {
	return sc.GetCertificateFn(ctx, vaultBaseURL, certificateName, certificateVersion)
}

func SetClientIDSecretInEnv() func() {
	oldClientID := os.Getenv("AZURE_CLIENT_ID")
	os.Setenv("AZURE_CLIENT_ID", "fake_client_id")
	oldClientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	os.Setenv("AZURE_CLIENT_SECRET", "fake_client_secret")
	return func() {
		os.Setenv("AZURE_CLIENT_ID", oldClientID)
		os.Setenv("AZURE_CLIENT_SECRET", oldClientSecret)
	}
}

func newJSONWebKey(b []byte) *keyvault.JSONWebKey {
	var key keyvault.JSONWebKey
	err := json.Unmarshal(b, &key)
	if err != nil {
		panic(err)
	}
	return &key
}
