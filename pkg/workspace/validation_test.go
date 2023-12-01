package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

func mockValidWorkspace() workspace.Workspace {
	return workspace.Workspace{
		Name:     "dev",
		Modules:  mockValidModuleConfigs(),
		Runtimes: mockValidRuntimeConfigs(),
		Backends: mockValidBackendConfigs(),
	}
}

func mockValidModuleConfigs() map[string]workspace.ModuleConfig {
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

func mockInvalidModuleConfigs() map[string]workspace.ModuleConfig {
	return map[string]workspace.ModuleConfig{
		"empty block name": {
			"": {},
		},
		"empty default block": {
			"default": {},
		},
		"not empty projectSelector in default block": {
			"default": {
				"type":            "aws",
				"version":         "5.7",
				"instanceType":    "db.t3.micro",
				"projectSelector": []string{"foo", "bar"},
			},
		},
		"empty patcher block": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"projectSelector": []string{"foo", "bar"},
			},
		},
		"empty project selector in patcher block": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"instanceType": "db.t3.small",
			},
		},
		"invalid project selector": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"instanceType":    "db.t3.small",
				"projectSelector": "foo",
			},
		},
		"empty projects": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"instanceType":    "db.t3.small",
				"projectSelector": []string{},
			},
		},
		"repeated projects in one patcher block": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"instanceType":    "db.t3.small",
				"projectSelector": []string{"foo", "foo"},
			},
		},
		"repeated projects in multiple patcher blocks": {
			"default": {
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			"smallClass": {
				"instanceType":    "db.t3.small",
				"projectSelector": []string{"foo"},
			},
			"largeClass": {
				"instanceType":    "db.t3.large",
				"projectSelector": []string{"foo"},
			},
		},
	}
}

func mockValidRuntimeConfigs() workspace.RuntimeConfigs {
	return workspace.RuntimeConfigs{
		Kubernetes: mockValidKubernetesConfig(),
		Terraform:  mockValidTerraformConfig(),
	}
}

func mockValidKubernetesConfig() workspace.KubernetesConfig {
	return workspace.KubernetesConfig{
		KubeConfig: "/etc/kubeconfig.yaml",
	}
}

func mockValidTerraformConfig() workspace.TerraformConfig {
	return workspace.TerraformConfig{
		"aws": {
			"version": "1.0.4",
			"source":  "hashicorp/aws",
			"region":  "us-east-1",
		},
	}
}

func mockValidBackendConfigs() workspace.BackendConfigs {
	return workspace.BackendConfigs{
		Local: mockValidLocalBackendConfig(),
	}
}

func mockValidLocalBackendConfig() workspace.LocalFileConfig {
	return workspace.LocalFileConfig{
		Path: "/etc/.kusion",
	}
}

func TestWorkspace_Validate(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		workspace workspace.Workspace
	}{
		{
			name:      "valid workspace",
			success:   true,
			workspace: mockValidWorkspace(),
		},
		{
			name:      "invalid workspace empty name",
			success:   false,
			workspace: workspace.Workspace{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWorkspace(&tc.workspace)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestModuleConfigs_Validate(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		moduleConfigs workspace.ModuleConfigs
	}{
		{
			name:          "valid module configs",
			success:       true,
			moduleConfigs: mockValidModuleConfigs(),
		},
		{
			name:    "invalid module configs empty module name",
			success: false,
			moduleConfigs: workspace.ModuleConfigs{
				"": mockValidModuleConfigs()["database"],
			},
		},
		{
			name:    "invalid module configs empty module config",
			success: false,
			moduleConfigs: workspace.ModuleConfigs{
				"database": nil,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateModuleConfigs(&tc.moduleConfigs)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestModuleConfig_Validate(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		moduleConfig workspace.ModuleConfig
	}{
		{
			name:         "valid module config",
			success:      true,
			moduleConfig: mockValidModuleConfigs()["database"],
		},
	}
	for desc, cfg := range mockInvalidModuleConfigs() {
		testcases = append(testcases, struct {
			name         string
			success      bool
			moduleConfig workspace.ModuleConfig
		}{
			name:         "invalid module config " + desc,
			success:      false,
			moduleConfig: cfg,
		})
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateModuleConfig(&tc.moduleConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestRuntimeConfigs_Validate(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		runtimeConfigs workspace.RuntimeConfigs
	}{
		{
			name:           "valid runtime configs",
			success:        true,
			runtimeConfigs: mockValidRuntimeConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRuntimeConfigs(&tc.runtimeConfigs)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestKubernetesConfig_Validate(t *testing.T) {
	testcases := []struct {
		name             string
		success          bool
		kubernetesConfig workspace.KubernetesConfig
	}{
		{
			name:             "valid kubernetes config",
			success:          true,
			kubernetesConfig: mockValidKubernetesConfig(),
		},
		{
			name:             "invalid kubernetes config empty kubeconfig",
			success:          false,
			kubernetesConfig: workspace.KubernetesConfig{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateKubernetesConfig(&tc.kubernetesConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestTerraformConfig_Validate(t *testing.T) {
	testcases := []struct {
		name            string
		success         bool
		terraformConfig workspace.TerraformConfig
	}{
		{
			name:            "valid terraform config",
			success:         true,
			terraformConfig: mockValidTerraformConfig(),
		},
		{
			name:    "invalid terraform config empty provider name",
			success: false,
			terraformConfig: workspace.TerraformConfig{
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
			terraformConfig: workspace.TerraformConfig{
				"aws": {},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTerraformConfig(&tc.terraformConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestBackendConfigs_Validate(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		backendConfigs workspace.BackendConfigs
	}{
		{
			name:           "valid backend configs",
			success:        true,
			backendConfigs: mockValidBackendConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBackendConfigs(&tc.backendConfigs)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestLocalBackendConfig_Validate(t *testing.T) {
	testcases := []struct {
		name               string
		success            bool
		localBackendConfig workspace.LocalFileConfig
	}{
		{
			name:               "valid local backend config",
			success:            true,
			localBackendConfig: mockValidLocalBackendConfig(),
		},
		{
			name:               "invalid local backend config empty path",
			success:            false,
			localBackendConfig: workspace.LocalFileConfig{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateLocalFileConfig(&tc.localBackendConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}
