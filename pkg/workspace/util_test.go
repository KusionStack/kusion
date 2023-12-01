package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

func mockModuleConfigs() map[string]workspace.ModuleConfig {
	return map[string]workspace.ModuleConfig{
		"database": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"instanceType":    "db.t3.small",
				"projectSelector": []string{"foo", "bar"},
			},
		},
		"port": {
			"default": {
				"type": "aws",
			},
		},
	}
}

func mockTerraformConfig() workspace.TerraformConfig {
	return workspace.TerraformConfig{
		"aws": {
			"version": "1.0.4",
			"source":  "hashicorp/aws",
			"region":  "us-east-1",
		},
	}
}

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
			moduleConfigs: mockModuleConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProjectModuleConfigs(&tc.moduleConfigs, tc.projectName)
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
			moduleConfig: mockModuleConfigs()["database"],
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
			moduleConfig: mockModuleConfigs()["database"],
		},
		{
			name:          "failed to get config empty project name",
			success:       false,
			projectName:   "",
			projectConfig: nil,
			moduleConfig:  mockModuleConfigs()["database"],
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProjectModuleConfig(&tc.moduleConfig, tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.projectConfig, cfg)
		})
	}
}

func Test_GetTerraformProviderConfig(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		providerName    string
		providerConfig  workspace.GenericConfig
		terraformConfig workspace.TerraformConfig
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
			terraformConfig: mockTerraformConfig(),
		},
		{
			name:            "failed to get config not exist provider",
			success:         false,
			providerName:    "alicloud",
			providerConfig:  nil,
			terraformConfig: mockTerraformConfig(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetTerraformProviderConfig(&tc.terraformConfig, tc.providerName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.providerConfig, cfg)
		})
	}
}
