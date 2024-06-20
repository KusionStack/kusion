package secretsmanager

import (
	"context"
	"fmt"
	"os"
	"strings"

	"kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/secrets"

	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/models"
	"github.com/aliyun/aliyun-secretsmanager-client-go/sdk/service"
	"github.com/tidwall/gjson"
)

const (
	errMissingProviderSpec     = "store spec is missing provider"
	errMissingAlicloudProvider = "invalid provider spec. Missing Alicloud field in store provider spec"
	errFailedToCreateClient    = "failed to create Alicloud Secrets Manager client: %w"
)

var (
	accessKeyID     = os.Getenv("credentials_access_key_id")
	accessKeySecret = os.Getenv("credentials_access_secret")
)

// DefaultSecretStoreProvider should implement the secrets.SecretStoreProvider interface.
var _ secrets.SecretStoreProvider = &DefaultSecretStoreProvider{}

// smSecretStore should implement the secrets.SecretStore interface.
var _ secrets.SecretStore = &smSecretStore{}

// DefaultSecretStoreProvider implements the secrets.SecretStoreProvider interface.
type DefaultSecretStoreProvider struct{}

// smSecretStore implements the secrets.SecretStore interface.
type smSecretStore struct {
	client Client
}

// NewSecretStore constructs a Vault based secret store with specific secret store spec.
func (p *DefaultSecretStoreProvider) NewSecretStore(spec *v1.SecretStore) (secrets.SecretStore, error) {
	providerSpec := spec.Provider
	if providerSpec == nil {
		return nil, fmt.Errorf(errMissingProviderSpec)
	}
	if providerSpec.Alicloud == nil {
		return nil, fmt.Errorf(errMissingAlicloudProvider)
	}

	client, err := getAlicloudClient(providerSpec.Alicloud.Region)
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateClient, err)
	}

	return &smSecretStore{
		client: client,
	}, nil
}

// getAlicloudClient returns an Alicloud Secrets Manager client with the specified region.
// Ref: https://github.com/aliyun/aliyun-secretsmanager-client-go/blob/v1.1.4/README.md
func getAlicloudClient(region string) (*sdk.SecretManagerCacheClient, error) {
	return sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().Standard().WithAccessKey(
			accessKeyID, accessKeySecret,
		).WithRegion(region).Build()).Build()
}

// GetSecret retrieves ref secret value from Alicloud Secrets Manager.
func (s *smSecretStore) GetSecret(ctx context.Context, ref v1.ExternalSecretRef) ([]byte, error) {
	secretInfo, err := s.client.GetSecretInfo(ref.Name)
	if err != nil {
		return nil, err
	}
	if ref.Property == "" {
		if secretInfo.SecretValue != "" {
			return []byte(secretInfo.SecretValue), nil
		}
		if secretInfo.SecretValueByteBuffer != nil {
			return secretInfo.SecretValueByteBuffer, nil
		}
		return nil, fmt.Errorf("invalid secret data. no secret value string nor binary for key: %s", ref.Name)
	}
	val := s.convertSecretToGjson(secretInfo, ref.Property)
	if !val.Exists() {
		return nil, fmt.Errorf("key %s does not exist in secret %s", ref.Property, ref.Name)
	}
	return []byte(val.String()), nil
}

func (s *smSecretStore) convertSecretToGjson(secretInfo *models.SecretInfo, refProperty string) gjson.Result {
	var payload string
	if secretInfo.SecretValue != "" {
		payload = secretInfo.SecretValue
	}
	if secretInfo.SecretValueByteBuffer != nil {
		payload = string(secretInfo.SecretValueByteBuffer)
	}

	// We need to search if a given key with a . exists before using gjson operations.
	idx := strings.Index(refProperty, ".")
	currentRefProperty := refProperty
	if idx > -1 {
		currentRefProperty = strings.ReplaceAll(refProperty, ".", "\\.")
		val := gjson.Get(payload, currentRefProperty)
		if !val.Exists() {
			currentRefProperty = refProperty
		}
	}

	return gjson.Get(payload, currentRefProperty)
}

func init() {
	secrets.Register(&DefaultSecretStoreProvider{}, &v1.ProviderSpec{
		Alicloud: &v1.AlicloudProvider{},
	})
}
