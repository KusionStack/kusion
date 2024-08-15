package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/constant"
)

const (
	SecretProviderTypeAliCloud   = "kam.v1.secretproviders.AlicloudProvider"
	SecretProviderTypeAWS        = "kam.v1.secretproviders.AWSProvider"
	SecretProviderTypeVault      = "kam.v1.secretproviders.VaultProvider"
	SecretProviderTypeAzureKV    = "kam.v1.secretproviders.AzureKVProvider"
	SecretProviderTypeOnPremises = "kam.v1.secretproviders.OnPremisesProvider"
)

// Workspace represents the specific configuration workspace
type Workspace struct {
	// ID is the id of the workspace.
	ID uint `yaml:"id" json:"id"`
	// Name is the name of the workspace.
	Name string `yaml:"name" json:"name"`
	// DisplayName is the human-readable display name.
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	// Description is a human-readable description of the workspace.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Labels are custom labels associated with the workspace.
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owners is a list of owners for the workspace.
	Owners []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// CreationTimestamp is the timestamp of the created for the workspace.
	CreationTimestamp time.Time `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	// UpdateTimestamp is the timestamp of the updated for the workspace.
	UpdateTimestamp time.Time `yaml:"updateTimestamp,omitempty" json:"updateTimestamp,omitempty"`
	// Backend is the corresponding backend for this workspace.
	Backend *Backend `yaml:"backend,omitempty" json:"backend,omitempty"`
}

type SecretValue struct {
	SecretStore *SecretStore `json:"secretStore"`
	//Value to store in the secret store.
	Value string `json:"value"`
	// Ref will only return with the update secret variable
	Ref string `json:"ref"`
}

type SecretStore struct {
	ProviderType string         `json:"providerType"`
	Provider     map[string]any `json:"provider"`
}

func (s *SecretStore) ConvertToKusionSecretStore() (*v1.SecretStore, error) {
	if s.ProviderType == "" {
		return nil, errors.New("miss provider type in secret store input")
	}
	rawProviderData, err := json.Marshal(s.Provider)
	if err != nil {
		return nil, err
	}
	switch s.ProviderType {
	case SecretProviderTypeAliCloud:
		aliCloudProvider := &v1.AlicloudProvider{}
		err = json.Unmarshal(rawProviderData, &aliCloudProvider)
		if err != nil {
			return nil, err
		}
		return &v1.SecretStore{Provider: &v1.ProviderSpec{Alicloud: aliCloudProvider}}, nil
	case SecretProviderTypeAWS:
		awsProvider := &v1.AWSProvider{}
		err = json.Unmarshal(rawProviderData, &awsProvider)
		if err != nil {
			return nil, err
		}
		return &v1.SecretStore{Provider: &v1.ProviderSpec{AWS: awsProvider}}, nil
	case SecretProviderTypeVault:
		vaultProvider := &v1.VaultProvider{}
		err = json.Unmarshal(rawProviderData, &vaultProvider)
		if err != nil {
			return nil, err
		}
		return &v1.SecretStore{Provider: &v1.ProviderSpec{Vault: vaultProvider}}, nil
	case SecretProviderTypeAzureKV:
		azureKVProvider := &v1.AzureKVProvider{}
		err = json.Unmarshal(rawProviderData, &azureKVProvider)
		if err != nil {
			return nil, err
		}
		return &v1.SecretStore{Provider: &v1.ProviderSpec{Azure: azureKVProvider}}, nil
	case SecretProviderTypeOnPremises:
		onPremisesProvider := &v1.OnPremisesProvider{}
		err = json.Unmarshal(rawProviderData, &onPremisesProvider)
		if err != nil {
			return nil, err
		}
		return &v1.SecretStore{Provider: &v1.ProviderSpec{OnPremises: onPremisesProvider}}, nil
	default:
		return nil, fmt.Errorf("illegal secret provider type %s", s.ProviderType)
	}
}

type WorkspaceFilter struct {
	BackendID uint
	Name      string
}

// Validate checks if the workspace is valid.
// It returns an error if the workspace is not valid.
func (w *Workspace) Validate() error {
	if w == nil {
		return constant.ErrWorkspaceNil
	}

	if w.Name == "" {
		return constant.ErrWorkspaceNameEmpty
	}

	if w.Backend == nil {
		return constant.ErrWorkspaceBackendNil
	}

	if len(w.Owners) == 0 {
		return constant.ErrWorkspaceOwnerNil
	}

	return nil
}
