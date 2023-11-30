package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockValidModuleConfigs() map[string]ModuleConfig {
	return map[string]ModuleConfig{
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

func mockInvalidModuleConfigs() map[string]ModuleConfig {
	return map[string]ModuleConfig{
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

func TestModuleConfigs_Validate(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		moduleConfigs ModuleConfigs
	}{
		{
			name:          "valid module configs",
			success:       true,
			moduleConfigs: mockValidModuleConfigs(),
		},
		{
			name:    "invalid module configs empty module name",
			success: false,
			moduleConfigs: ModuleConfigs{
				"": mockValidModuleConfigs()["database"],
			},
		},
		{
			name:    "invalid module configs empty module config",
			success: false,
			moduleConfigs: ModuleConfigs{
				"database": nil,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.moduleConfigs.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestModuleConfigs_GetProjectModuleConfigs(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		projectName    string
		projectConfigs map[string]GenericConfig
		moduleConfigs  ModuleConfigs
	}{
		{
			name:        "successfully get project module configs",
			success:     true,
			projectName: "foo",
			projectConfigs: map[string]GenericConfig{
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
			cfg, err := tc.moduleConfigs.GetProjectModuleConfigs(tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.projectConfigs, cfg)
		})
	}
}

func TestModuleConfig_Validate(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		moduleConfig ModuleConfig
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
			moduleConfig ModuleConfig
		}{
			name:         "invalid module config " + desc,
			success:      false,
			moduleConfig: cfg,
		})
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.moduleConfig.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestModuleConfig_GetProjectModuleConfig(t *testing.T) {
	testcases := []struct {
		name          string
		success       bool
		projectName   string
		projectConfig GenericConfig
		moduleConfig  ModuleConfig
	}{
		{
			name:        "successfully get default project module config",
			success:     true,
			projectName: "baz",
			projectConfig: GenericConfig{
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
			projectConfig: GenericConfig{
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
			cfg, err := tc.moduleConfig.GetProjectModuleConfig(tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.projectConfig, cfg)
		})
	}
}
