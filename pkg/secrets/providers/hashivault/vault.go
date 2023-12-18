package hashivault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/tidwall/gjson"

	secretsapi "kusionstack.io/kusion/pkg/apis/secrets"
	"kusionstack.io/kusion/pkg/secrets"
)

const (
	errInvalidVaultSecretStore = "cannot find valid Vault provider spec"
	errReadSecret              = "failed to read secret data from Vault: %w"
	errParseDataField          = "failed to find data field"
	errJSONUnmarshall          = "failed to unmarshall JSON"
	errUnexpectedKey           = "unexpected key in secret data: %s"
	errDataPropertyFormat      = "unexpected data format %s for property field: %s"
	errSecretFormat            = "cannot find property %s in secret data"
	errBuildVaultClient        = "failed to new Vault client: %w"
)

// DefaultFactory should implement the secrets.SecretStoreFactory interface
var _ secrets.SecretStoreFactory = &DefaultFactory{}

// vaultSecretStore should implement the secrets.SecretStore interface
var _ secrets.SecretStore = &vaultSecretStore{}

type DefaultFactory struct{}

func (p *DefaultFactory) Type() string {
	return "Vault"
}

// NewSecretStore constructs a Vault based secret store with specific secret store spec.
func (p *DefaultFactory) NewSecretStore(spec secretsapi.SecretStoreSpec) (secrets.SecretStore, error) {
	providerSpec := spec.Provider
	if providerSpec == nil || providerSpec.Vault == nil {
		return nil, errors.New(errInvalidVaultSecretStore)
	}

	vaultSpec := providerSpec.Vault
	client, err := getVaultClient(vaultSpec.Server)
	if err != nil {
		return nil, err
	}

	store := vaultSecretStore{
		provider: vaultSpec,
		logical:  client.Logical(),
	}
	return &store, nil
}

func getVaultClient(server string) (*vault.Client, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = server
	c, err := vault.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf(errBuildVaultClient, err)
	}
	token := getVaultToken()
	if token != "" {
		c.SetToken(token)
	}
	return c, nil
}

// getVaultToken ensures that we check both VAULT_SERVER_TOKEN and VAULT_TOKEN environment
// variables for the API token for vault. VAULT_SERVER_TOKEN takes precedence over VAULT_TOKEN.
// If neither environment variables are found, then we return an empty string as token is not required.
func getVaultToken() string {
	serverToken := os.Getenv("VAULT_SERVER_TOKEN")
	if serverToken != "" {
		return serverToken
	}

	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken != "" {
		return vaultToken
	}

	return ""
}

type vaultSecretStore struct {
	provider *secretsapi.VaultProvider
	logical  Logical
}

// GetSecret retrieves ref secret value from Vault server.
func (v *vaultSecretStore) GetSecret(ctx context.Context, ref secretsapi.ExternalSecretRef) ([]byte, error) {
	secretData, err := v.readSecret(ctx, ref.Name, ref.Version)
	if err != nil {
		return nil, err
	}
	jsonStr, err := json.Marshal(secretData)
	if err != nil {
		return nil, err
	}
	// return raw json if no property is defined
	if ref.Property == "" {
		return jsonStr, nil
	}

	// First try to extract key from secret with raw property
	if _, ok := secretData[ref.Property]; ok {
		return getTypedKey(secretData, ref.Property)
	}

	// Then extract key from secret using gjson lib
	val := gjson.Get(string(jsonStr), ref.Property)
	if !val.Exists() {
		return nil, fmt.Errorf(errSecretFormat, ref.Property)
	}
	return []byte(val.String()), nil
}

func (v *vaultSecretStore) readSecret(ctx context.Context, path, version string) (map[string]interface{}, error) {
	// build correct path according to vault docs for v1 and v2 API
	secretPath := v.buildPath(path)
	var params map[string][]string
	if version != "" {
		params = make(map[string][]string)
		params["version"] = []string{version}
	}
	secret, err := v.logical.ReadWithDataWithContext(ctx, secretPath, params)
	if err != nil {
		return nil, fmt.Errorf(errReadSecret, err)
	}
	if secret == nil {
		// return empty secret data
		return map[string]interface{}{}, nil
	}
	secretData := secret.Data
	if v.provider.Version == secretsapi.VaultKVStoreV2 {
		// Vault KV2 has data embedded within sub-field
		// Ref: https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2#read-secret-version
		embeddedData, ok := secretData["data"]
		if !ok {
			return nil, errors.New(errParseDataField)
		}
		if embeddedData == nil {
			// return empty secret data
			return map[string]interface{}{}, nil
		}
		secretData, ok = embeddedData.(map[string]interface{})
		if !ok {
			return nil, errors.New(errJSONUnmarshall)
		}
	}
	return secretData, nil
}

// buildPath is a helper method to build the final secret path. The path build logic
// varies depending on the Vault KV secrets engine version:
// v1: https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v1#read-secret
// v2: https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2#read-secret-version
func (v *vaultSecretStore) buildPath(path string) string {
	mountPath := v.provider.Path
	out := path
	if mountPath != nil {
		prefixToCut := *mountPath + "/"
		if strings.HasPrefix(out, prefixToCut) {
			_, out, _ = strings.Cut(out, prefixToCut)
			// if data succeeds mountPath on v2 store, we should remove it as well
			if strings.HasPrefix(out, "data/") && v.provider.Version == secretsapi.VaultKVStoreV2 {
				_, out, _ = strings.Cut(out, "data/")
			}
		}
		buildPath := strings.Split(out, "/")
		buildMount := strings.Split(*mountPath, "/")
		if v.provider.Version == secretsapi.VaultKVStoreV2 {
			buildMount = append(buildMount, "data")
		}
		buildMount = append(buildMount, buildPath...)
		out = strings.Join(buildMount, "/")
		return out
	}
	if !strings.Contains(out, "/data/") && v.provider.Version == secretsapi.VaultKVStoreV2 {
		buildPath := strings.Split(out, "/")
		buildMount := []string{buildPath[0], "data"}
		buildMount = append(buildMount, buildPath[1:]...)
		out = strings.Join(buildMount, "/")
		return out
	}
	return out
}

func getTypedKey(data map[string]interface{}, key string) ([]byte, error) {
	v, ok := data[key]
	if !ok {
		return nil, fmt.Errorf(errUnexpectedKey, key)
	}
	switch t := v.(type) {
	case string:
		return []byte(t), nil
	case map[string]interface{}:
		return json.Marshal(t)
	case []string:
		return []byte(strings.Join(t, "\n")), nil
	case []byte:
		return t, nil
	// also covers int and float32 due to json.Marshal
	case float64:
		return []byte(strconv.FormatFloat(t, 'f', -1, 64)), nil
	case json.Number:
		return []byte(t.String()), nil
	case []interface{}:
		return json.Marshal(t)
	case bool:
		return []byte(strconv.FormatBool(t)), nil
	case nil:
		return []byte(nil), nil
	default:
		return nil, fmt.Errorf(errDataPropertyFormat, key, reflect.TypeOf(t))
	}
}

func init() {
	secrets.Register(&DefaultFactory{}, &secretsapi.ProviderSpec{
		Vault: &secretsapi.VaultProvider{},
	})
}
