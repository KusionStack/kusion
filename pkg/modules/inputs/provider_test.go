package inputs

import (
	"testing"

	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
)

func TestSetString(t *testing.T) {
	provider := &Provider{}
	if err := provider.SetString("registry.terraform.io/hashicorp/aws/5.0.1"); err != nil {
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

	expectedURL := "registry.terraform.io/hashicorp/aws/5.0.1"
	if url != expectedURL {
		t.Errorf("unexpected url, got: %s, expected: %s", url, expectedURL)
	}
}

func TestGetProviderRegion(t *testing.T) {
	providerConfig := &workspaceapi.ProviderConfig{
		Source:  "hashicorp/aws",
		Version: "5.0.1",
		GenericConfig: workspaceapi.GenericConfig{
			"region": "us-east-1",
		},
	}

	region, err := GetProviderRegion(providerConfig)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedRegion := "us-east-1"
	if region != expectedRegion {
		t.Errorf("unexpected region, got: %s, expected: %s", region, expectedRegion)
	}
}
