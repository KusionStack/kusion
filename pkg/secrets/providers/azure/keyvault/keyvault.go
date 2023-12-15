package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/tidwall/gjson"

	secretsapi "kusionstack.io/kusion/pkg/apis/secrets"
	"kusionstack.io/kusion/pkg/secrets"
)

const (
	defaultObjType           = "secret"
	objectTypeCert           = "cert"
	objectTypeKey            = "key"
	errMissingProviderSpec   = "store spec is missing provider"
	errMissingAzureProvider  = "invalid provider spec. Missing Azure field in store provider spec"
	errMissingTenant         = "missing tenantID in store provider spec"
	errMissingClientIDSecret = "cannot read clientID/clientSecret from environment variables"
	errPropertyNotExist      = "property %s does not exist in key %s"
	errUnknownObjectType     = "unknown Azure KeyVault object Type for %s"
)

// DefaultFactory should implement the secrets.SecretStoreFactory interface
var _ secrets.SecretStoreFactory = &DefaultFactory{}

// kvSecretStore should implement the secrets.SecretStore interface
var _ secrets.SecretStore = &kvSecretStore{}

type DefaultFactory struct{}

// NewSecretStore constructs an Azure KeyVault based secret store with specific secret store spec.
func (p *DefaultFactory) NewSecretStore(spec secretsapi.SecretStoreSpec) (secrets.SecretStore, error) {
	providerSpec := spec.Provider
	if providerSpec == nil {
		return nil, fmt.Errorf(errMissingProviderSpec)
	}
	if providerSpec.Azure == nil {
		return nil, fmt.Errorf(errMissingAzureProvider)
	}

	secretClient, err := getSecretClient(providerSpec.Azure)
	if err != nil {
		return nil, err
	}
	return &kvSecretStore{secretClient, providerSpec.Azure}, nil
}

func getSecretClient(spec *secretsapi.AzureKVProvider) (SecretClient, error) {
	authorizer, err := authorizerForServicePrincipal(spec)
	if err != nil {
		return nil, err
	}
	client := keyvault.New()
	client.Authorizer = authorizer
	return client, nil
}

// authorizerForServicePrincipal returns a service principal based authorizer used by clients to access to Azure.
// By-default it uses credentials from the environment;
// See https://docs.microsoft.com/en-us/go/azure/azure-sdk-go-authorization#use-environment-based-authentication.
func authorizerForServicePrincipal(spec *secretsapi.AzureKVProvider) (autorest.Authorizer, error) {
	if spec.TenantID == nil {
		return nil, fmt.Errorf(errMissingTenant)
	}

	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf(errMissingClientIDSecret)
	}

	clientCredentialsConfig := auth.NewClientCredentialsConfig(clientID, clientSecret, *spec.TenantID)
	clientCredentialsConfig.Resource = kvResourceForProviderConfig(spec.EnvironmentType)
	clientCredentialsConfig.AADEndpoint = adEndpointForEnvironmentType(spec.EnvironmentType)
	return clientCredentialsConfig.Authorizer()
}

func adEndpointForEnvironmentType(t secretsapi.AzureEnvironmentType) string {
	switch t {
	case secretsapi.AzureEnvironmentPublicCloud:
		return azure.PublicCloud.ActiveDirectoryEndpoint
	case secretsapi.AzureEnvironmentChinaCloud:
		return azure.ChinaCloud.ActiveDirectoryEndpoint
	case secretsapi.AzureEnvironmentUSGovernmentCloud:
		return azure.USGovernmentCloud.ActiveDirectoryEndpoint
	case secretsapi.AzureEnvironmentGermanCloud:
		return azure.GermanCloud.ActiveDirectoryEndpoint
	default:
		return azure.PublicCloud.ActiveDirectoryEndpoint
	}
}

func kvResourceForProviderConfig(t secretsapi.AzureEnvironmentType) string {
	var res string
	switch t {
	case secretsapi.AzureEnvironmentPublicCloud:
		res = azure.PublicCloud.KeyVaultEndpoint
	case secretsapi.AzureEnvironmentChinaCloud:
		res = azure.ChinaCloud.KeyVaultEndpoint
	case secretsapi.AzureEnvironmentUSGovernmentCloud:
		res = azure.USGovernmentCloud.KeyVaultEndpoint
	case secretsapi.AzureEnvironmentGermanCloud:
		res = azure.GermanCloud.KeyVaultEndpoint
	default:
		res = azure.PublicCloud.KeyVaultEndpoint
	}
	return strings.TrimSuffix(res, "/")
}

type kvSecretStore struct {
	secretClient SecretClient
	provider     *secretsapi.AzureKVProvider
}

// GetSecret retrieves ref secret value from Azure KeyVault.
func (k *kvSecretStore) GetSecret(ctx context.Context, ref secretsapi.ExternalSecretRef) ([]byte, error) {
	objectType, secretName := getObjType(ref)

	switch objectType {
	case defaultObjType:
		// returns a SecretBundle with the secret value
		// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault#SecretBundle
		secretResp, err := k.secretClient.GetSecret(ctx, *k.provider.VaultURL, secretName, ref.Version)
		if err != nil {
			return nil, err
		}
		return getProperty(*secretResp.Value, ref.Property, ref.Name)
	case objectTypeCert:
		// returns a CertBundle. We return CER contents of x509 certificate
		// see: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault#CertificateBundle
		certResp, err := k.secretClient.GetCertificate(ctx, *k.provider.VaultURL, secretName, ref.Version)
		if err != nil {
			return nil, err
		}
		return *certResp.Cer, nil
	case objectTypeKey:
		// returns a KeyBundle that contains a WebKey
		// see: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault#KeyBundle
		keyResp, err := k.secretClient.GetKey(ctx, *k.provider.VaultURL, secretName, ref.Version)
		if err != nil {
			return nil, err
		}
		return json.Marshal(keyResp.Key)
	}

	return nil, fmt.Errorf(errUnknownObjectType, secretName)
}

// Retrieves a property value if specified and the secret value if not.
func getProperty(secret, property, key string) ([]byte, error) {
	if property == "" {
		return []byte(secret), nil
	}
	res := gjson.Get(secret, property)
	if !res.Exists() {
		idx := strings.Index(property, ".")
		if idx < 0 {
			return nil, fmt.Errorf(errPropertyNotExist, property, key)
		}
		escaped := strings.ReplaceAll(property, ".", "\\.")
		jValue := gjson.Get(secret, escaped)
		if jValue.Exists() {
			return []byte(jValue.String()), nil
		}
		return nil, fmt.Errorf(errPropertyNotExist, property, key)
	}
	return []byte(res.String()), nil
}

func getObjType(ref secretsapi.ExternalSecretRef) (string, string) {
	objectType := defaultObjType

	secretName := ref.Name
	nameSlice := strings.Split(ref.Name, "/")

	if len(nameSlice) > 1 {
		objectType = nameSlice[0]
		secretName = nameSlice[1]
	}
	return objectType, secretName
}
