package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

func Test_GetProjectModuleConfigs(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		projectName    string
		projectConfigs map[string]workspace.GenericConfig
		moduleConfigs  workspace.ModuleConfigs
	}{
		{
			name:        "successfully get project module configs",
			success:     true,
			projectName: "foo",
			projectConfigs: map[string]workspace.GenericConfig{
				"database": {
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.small",
				},
				"port": {
					"type": "aws",
				},
			},
			moduleConfigs: mockValidModuleConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProjectModuleConfigs(tc.moduleConfigs, tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.projectConfigs, cfg)
		})
	}
}

func Test_GetProjectModuleConfig(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		projectName   string
		projectConfig workspace.GenericConfig
		moduleConfig  workspace.ModuleConfig
	}{
		{
			name:        "successfully get default project module config",
			success:     true,
			projectName: "baz",
			projectConfig: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			moduleConfig: mockValidModuleConfigs()["database"],
		},
		{
			name:        "successfully get override project module config",
			success:     true,
			projectName: "foo",
			projectConfig: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.small",
			},
			moduleConfig: mockValidModuleConfigs()["database"],
		},
		{
			name:          "failed to get config empty project name",
			success:       false,
			projectName:   "",
			projectConfig: nil,
			moduleConfig:  mockValidModuleConfigs()["database"],
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProjectModuleConfig(tc.moduleConfig, tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.projectConfig, cfg)
		})
	}
}

func Test_GetKubernetesConfig(t *testing.T) {
	testcases := []struct {
		name             string
		success          bool
		kubernetesConfig *workspace.KubernetesConfig
		runtimeConfigs   *workspace.RuntimeConfigs
	}{
		{
			name:             "successfully get kubernetes config",
			success:          true,
			kubernetesConfig: mockValidKubernetesConfig(),
			runtimeConfigs:   mockValidRuntimeConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetKubernetesConfig(tc.runtimeConfigs)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.kubernetesConfig, cfg)
		})
	}
}

func Test_GetTerraformProviderConfig(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		providerName   string
		providerConfig workspace.GenericConfig
		runtimeConfigs *workspace.RuntimeConfigs
	}{
		{
			name:         "successfully get terraform provider config",
			success:      true,
			providerName: "aws",
			providerConfig: workspace.GenericConfig{
				"version": "1.0.4",
				"source":  "hashicorp/aws",
				"region":  "us-east-1",
			},
			runtimeConfigs: mockValidRuntimeConfigs(),
		},
		{
			name:           "failed to get config not exist provider",
			success:        false,
			providerName:   "alicloud",
			providerConfig: nil,
			runtimeConfigs: mockValidRuntimeConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetTerraformProviderConfig(tc.runtimeConfigs, tc.providerName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.providerConfig, cfg)
		})
	}
}
