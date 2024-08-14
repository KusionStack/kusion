package secretsmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/google/uuid"

	"kusionstack.io/kusion/pkg/secrets/providers/viettelcloud/vclient"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"

	"kusionstack.io/kusion/pkg/secrets"
)

const (
	errMissingProviderSpec         = "store spec is missing provider"
	errMissingViettelCloudProvider = "invalid provider spec. Missing ViettelCloud field in store provider spec"
	errFailedToCreateClient        = "failed to create ViettelCloud Secrets Manager client: %w"
)

var (
	viettelcloudCmpURL    = os.Getenv("VIETTEL_CLOUD_CMP_URL")
	viettelcloudUserToken = os.Getenv("VIETTEL_CLOUD_USER_TOKEN")
	viettelcloudProjectID = os.Getenv("VIETTEL_CLOUD_PROJECT_ID")
)

// DefaultSecretStoreProvider should implement the secrets.SecretStoreProvider interface.
var _ secrets.SecretStoreProvider = &DefaultSecretStoreProvider{}

// smSecretStore should implement the secrets.SecretStore interface.
var _ secrets.SecretStore = &smSecretStore{}

// DefaultSecretStoreProvider implements the secrets.SecretStoreProvider interface.
type DefaultSecretStoreProvider struct{}

// smSecretStore implements the secrets.SecretStore interface.
type smSecretStore struct {
	client    Client
	projectID uuid.UUID
}

func (p *DefaultSecretStoreProvider) NewSecretStore(spec *v1.SecretStore) (secrets.SecretStore, error) {
	providerSpec := spec.Provider
	if providerSpec == nil {
		return nil, fmt.Errorf(errMissingProviderSpec)
	}
	if providerSpec.ViettelCloud == nil {
		return nil, fmt.Errorf(errMissingViettelCloudProvider)
	}
	var project string
	if providerSpec.ViettelCloud.ProjectID != "" {
		project = providerSpec.ViettelCloud.ProjectID
	} else {
		project = viettelcloudProjectID
	}

	projectID, err := uuid.Parse(project)
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateClient, err)
	}

	var cmpURL string
	if providerSpec.ViettelCloud.CmpURL != "" {
		cmpURL = providerSpec.ViettelCloud.CmpURL
	} else {
		cmpURL = viettelcloudCmpURL
	}

	patToken, err := vclient.NewSecurityProviderPATToken(viettelcloudUserToken)
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateClient, err)
	}
	client, err := vclient.NewClientWithResponses(cmpURL, vclient.WithRequestEditorFn(patToken.Intercept))
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateClient, err)
	}
	return &smSecretStore{
		client:    client,
		projectID: projectID,
	}, nil
}

func (s *smSecretStore) GetSecret(ctx context.Context, ref v1.ExternalSecretRef) ([]byte, error) {
	secretResponse, err := s.client.SecretManagerSecretsRetrieveWithResponse(ctx, ref.Name, &vclient.SecretManagerSecretsRetrieveParams{
		ProjectID: s.projectID,
	})
	if err != nil {
		return nil, err
	}
	if secretResponse.JSON200 == nil {
		return nil, fmt.Errorf("response body is nil")
	}
	if secretResponse.JSON200.Secret == nil {
		return nil, fmt.Errorf("secret is nil")
	}
	if secretResponse.JSON400 != nil {
		return nil, fmt.Errorf("error response: %s", secretResponse.JSON400.Union)
	}
	if secretResponse.JSON401 != nil {
		return nil, fmt.Errorf("error response: %s", secretResponse.JSON401.Union)
	}
	if secretResponse.JSON403 != nil {
		return nil, fmt.Errorf("error response: %s", secretResponse.JSON403.Union)
	}
	val := s.convertSecretToGjson(secretResponse.JSON200, ref.Property)
	if !val.Exists() {
		return nil, fmt.Errorf("key %s does not exist in secret %s", ref.Property, ref.Name)
	}
	return []byte(val.String()), nil
}

func (s *smSecretStore) convertSecretToGjson(secretInfo *vclient.SecretRetrieve, refProperty string) gjson.Result {
	var payload string
	if secretInfo.Secret != nil {
		// todo: handle nested secrets
		data, err := json.Marshal(*secretInfo.Secret)
		if err != nil {
			return gjson.Result{}
		}
		payload = string(data)
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
		ViettelCloud: &v1.ViettelCloudProvider{},
	})
}
