package inputs

import (
	"testing"

	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
)

var (
	awsProviderURL = "registry.terraform.io/hashicorp/aws/5.0.1"
)

func TestSetString(t *testing.T) {
	provider := &Provider{}
	if err := provider.SetString(awsProviderURL); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedURL := "registry.terraform.io/hashicorp/aws/5.0.1"
	expectedHost := "registry.terraform.io"
	expectedNamespace := "hashicorp"
	expectedName := "aws"
	expectedVersion := "5.0.1"

	if provider.URL != expectedURL {
		t.Errorf("unexpected url, got: %s, expected: %s", provider.URL, expectedURL)
	}
	if provider.Host != expectedHost {
		t.Errorf("unexpected host, got: %s, expected: %s", provider.Host, expectedHost)
	}
	if provider.Namespace != expectedNamespace {
		t.Errorf("unexpected namespace, got: %s, expected: %s", provider.Namespace, expectedNamespace)
	}
	if provider.Name != expectedName {
		t.Errorf("unexpected name, got: %s, expected: %s", provider.Name, expectedName)
	}
	if provider.Version != expectedVersion {
		t.Errorf("unexpected version, got: %s, expected: %s", provider.Version, expectedVersion)
	}
}

func TestGetProviderURL(t *testing.T) {
	providerConfig := &workspaceapi.ProviderConfig{
		Source:  "hashicorp/aws",
		Version: "5.0.1",
	}

	url, err := GetProviderURL(providerConfig)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedURL := awsProviderURL
	if url != expectedURL {
		t.Errorf("unexpected url, got: %s, want: %s", url, expectedURL)
	}
}
