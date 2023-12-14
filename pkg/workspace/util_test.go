package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/container"
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
		moduleConfig  *workspace.ModuleConfig
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

func Test_FormatGenericConfig(t *testing.T) {
	testcases := []struct {
		name        string
		success     bool
		src         workspace.GenericConfig
		dst         any
		expectedDst any
	}{
		{
			name:    "successfully format",
			success: true,
			src: workspace.GenericConfig{
				"containers": map[string]any{
					"hello": map[string]any{
						"resources": map[string]any{
							"cpu":    "2",
							"memory": "2Gi",
						},
						"files": map[string]any{
							"test-file": map[string]any{
								"content": "test-content",
								"mode":    "0644",
							},
						},
						"livenessProbe": map[string]any{
							"probeHandler": map[string]any{
								"_type":   "Exec",
								"command": []any{"cat", "/tmp/health"},
							},
							"initialDelaySeconds": 30,
						},
						"readinessProbe": map[string]any{
							"probeHandler": map[string]any{
								"_type": "Http",
								"url":   "http://localhost:80",
								"headers": map[string]any{
									"header-k1": "header-v1",
								},
							},
						},
					},
				},
				"replicas": 2,
				"labels": map[string]any{
					"label-k1": "label-v1",
				},
				"type": "Collaset",
			},
			dst: &workload.Service{},
			expectedDst: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"hello": {
							Resources: map[string]string{
								"cpu":    "2",
								"memory": "2Gi",
							},
							Files: map[string]container.FileSpec{
								"test-file": {
									Content: "test-content",
									Mode:    "0644",
								},
							},
							LivenessProbe: &container.Probe{
								ProbeHandler: &container.ProbeHandler{
									TypeWrapper: container.TypeWrapper{
										Type: "Exec",
									},
									ExecAction: &container.ExecAction{
										Command: []string{"cat", "/tmp/health"},
									},
								},
								InitialDelaySeconds: 30,
							},
							ReadinessProbe: &container.Probe{
								ProbeHandler: &container.ProbeHandler{
									TypeWrapper: container.TypeWrapper{
										Type: "Http",
									},
									HTTPGetAction: &container.HTTPGetAction{
										URL: "http://localhost:80",
										Headers: map[string]string{
											"header-k1": "header-v1",
										},
									},
								},
							},
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"label-k1": "label-v1",
					},
				},
				Type: "Collaset",
			},
		},
		{
			name:        "failed to format not initialized dst",
			success:     false,
			src:         workspace.GenericConfig{},
			dst:         nil,
			expectedDst: nil,
		},
		{
			name:        "failed to format not pointer of struct dst",
			success:     false,
			src:         workspace.GenericConfig{},
			dst:         "invalid type",
			expectedDst: nil,
		},
		{
			name:    "format failed not empty dst",
			success: false,
			src:     workspace.GenericConfig{},
			dst: &workload.Service{
				Type: "Collaset",
			},
			expectedDst: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := FormatGenericConfig(tc.src, tc.dst)
			assert.Equal(t, tc.success, err == nil)
			if err == nil {
				assert.Equal(t, tc.expectedDst, tc.dst)
			}
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
		providerConfig *workspace.ProviderConfig
		runtimeConfigs *workspace.RuntimeConfigs
	}{
		{
			name:         "successfully get terraform provider config",
			success:      true,
			providerName: "aws",
			providerConfig: &workspace.ProviderConfig{
				Source:  "hashicorp/aws",
				Version: "1.0.4",
				GenericConfig: workspace.GenericConfig{
					"region": "us-east-1",
				},
			},
			runtimeConfigs: mockValidRuntimeConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProviderConfig(tc.runtimeConfigs, tc.providerName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.providerConfig, cfg)
		})
	}
}
