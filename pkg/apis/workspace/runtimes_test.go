package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockValidRuntimeConfigs() RuntimeConfigs {
	return RuntimeConfigs{
		Kubernetes: mockValidKubernetesConfig(),
		Terraform:  mockValidTerraformConfig(),
	}
}

func mockValidKubernetesConfig() KubernetesConfig {
	return KubernetesConfig{
		KubeConfig: "/etc/kubeconfig.yaml",
	}
}

func mockValidTerraformConfig() TerraformConfig {
	return TerraformConfig{
		"aws": {
			"version": "1.0.4",
			"source":  "hashicorp/aws",
			"region":  "us-east-1",
		},
	}
}

func TestRuntimeConfigs_Validate(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		runtimeConfigs RuntimeConfigs
	}{
		{
			name:           "valid runtime configs",
			success:        true,
			runtimeConfigs: mockValidRuntimeConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.runtimeConfigs.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestKubernetesConfig_Validate(t *testing.T) {
	testcases := []struct {
		name             string
		success          bool
		kubernetesConfig KubernetesConfig
	}{
		{
			name:             "valid kubernetes config",
			success:          true,
			kubernetesConfig: mockValidKubernetesConfig(),
		},
		{
			name:             "invalid kubernetes config empty kubeconfig",
			success:          false,
			kubernetesConfig: KubernetesConfig{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.kubernetesConfig.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestTerraformConfig_Validate(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		terraformConfig TerraformConfig
	}{
		{
			name:            "valid terraform config",
			success:         true,
			terraformConfig: mockValidTerraformConfig(),
		},
		{
			name:    "invalid terraform config empty provider name",
			success: false,
			terraformConfig: TerraformConfig{
				"": {
					"version": "1.0.4",
					"source":  "hashicorp/aws",
					"region":  "us-east-1",
				},
			},
		},
		{
			name:    "invalid terraform config empty provider config",
			success: false,
			terraformConfig: TerraformConfig{
				"aws": {},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.terraformConfig.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestTerraformConfig_GetProviderConfig(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		providerName    string
		providerConfig  GenericConfig
		terraformConfig TerraformConfig
	}{
		{
			name:         "successfully get terraform provider config",
			success:      true,
			providerName: "aws",
			providerConfig: GenericConfig{
				"version": "1.0.4",
				"source":  "hashicorp/aws",
				"region":  "us-east-1",
			},
			terraformConfig: mockValidTerraformConfig(),
		},
		{
			name:            "failed to get config not exist provider",
			success:         false,
			providerName:    "alicloud",
			providerConfig:  nil,
			terraformConfig: mockValidTerraformConfig(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := tc.terraformConfig.GetProviderConfig(tc.providerName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.providerConfig, cfg)
		})
	}
}
