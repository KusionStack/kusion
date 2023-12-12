package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/workspace"
)

func mockValidWorkspace(name string) *workspace.Workspace {
	return &workspace.Workspace{
		Name:     name,
		Modules:  mockValidModuleConfigs(),
		Runtimes: mockValidRuntimeConfigs(),
		Backends: mockValidBackendConfigs(),
	}
}

func mockValidModuleConfigs() map[string]*workspace.ModuleConfig {
	return map[string]*workspace.ModuleConfig{
		"database": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "bar"},
				},
			},
		},
		"port": {
			Default: workspace.GenericConfig{
				"type": "aws",
			},
		},
	}
}

func mockInvalidModuleConfigs() map[string]workspace.ModuleConfig {
	return map[string]workspace.ModuleConfig{
		"empty default block": {
			Default: workspace.GenericConfig{},
		},
		"not empty projectSelector in default block": {
			Default: workspace.GenericConfig{
				"type":            "aws",
				"version":         "5.7",
				"instanceType":    "db.t3.micro",
				"projectSelector": []string{"foo", "bar"},
			},
		},
		"empty patcher block name": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "bar"},
				},
			},
		},
		"empty patcher block": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": nil,
			},
		},
		"empty config in patcher block": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": {
					ProjectSelector: []string{"foo", "bar"},
				},
			},
		},
		"empty project selector in patcher block": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.small",
					},
				},
			},
		},
		"empty project name": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"", "bar"},
				},
			},
		},
		"repeated projects in one patcher block": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "foo"},
				},
			},
		},
		"repeated projects in multiple patcher blocks": {
			Default: workspace.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: workspace.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "bar"},
				},
				"largeClass": {
					GenericConfig: workspace.GenericConfig{
						"instanceType": "db.t3.large",
					},
					ProjectSelector: []string{"foo"},
				},
			},
		},
	}
}

func mockValidRuntimeConfigs() *workspace.RuntimeConfigs {
	return &workspace.RuntimeConfigs{
		Kubernetes: mockValidKubernetesConfig(),
		Terraform:  mockValidTerraformConfig(),
	}
}

func mockValidKubernetesConfig() *workspace.KubernetesConfig {
	return &workspace.KubernetesConfig{
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

func mockValidBackendConfigs() *workspace.BackendConfigs {
	return &workspace.BackendConfigs{
		Local: &workspace.LocalFileConfig{},
	}
}

func mockValidMysqlConfig() *workspace.MysqlConfig {
	return &workspace.MysqlConfig{
		DBName: "kusion_db",
		User:   "kusion",
		Host:   "127.0.0.1",
	}
}

func mockValidGenericObjectStorageConfig() *workspace.GenericObjectStorageConfig {
	return &workspace.GenericObjectStorageConfig{
		Bucket: "kusion_bucket",
	}
}

func mockValidCompletedOssConfig() *workspace.OssConfig {
	return &workspace.OssConfig{
		GenericObjectStorageConfig: workspace.GenericObjectStorageConfig{
			Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
			AccessKeyID:     "fake-access-key-id",
			AccessKeySecret: "fake-access-key-secret",
			Bucket:          "kusion_bucket",
		},
	}
}

func mockValidCompletedS3Config() *workspace.S3Config {
	return &workspace.S3Config{
		GenericObjectStorageConfig: workspace.GenericObjectStorageConfig{
			AccessKeyID:     "fake-access-key-id",
			AccessKeySecret: "fake-access-key-secret",
			Bucket:          "kusion_bucket",
		},
		Region: "us-east-1",
	}
}

func TestValidateWorkspace(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		workspace *workspace.Workspace
	}{
		{
			name:      "valid workspace",
			success:   true,
			workspace: mockValidWorkspace("dev"),
		},
		{
			name:      "invalid workspace empty name",
			success:   false,
			workspace: &workspace.Workspace{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWorkspace(tc.workspace)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateModuleConfigs(t *testing.T) {
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
			err := ValidateModuleConfigs(tc.moduleConfigs)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateModuleConfig(t *testing.T) {
	testcases := []struct {
		name         string
		success      bool
		moduleConfig workspace.ModuleConfig
	}{
		{
			name:         "valid module config",
			success:      true,
			moduleConfig: *mockValidModuleConfigs()["database"],
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

func TestValidateRuntimeConfigs(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		runtimeConfigs *workspace.RuntimeConfigs
	}{
		{
			name:           "valid runtime configs",
			success:        true,
			runtimeConfigs: mockValidRuntimeConfigs(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRuntimeConfigs(tc.runtimeConfigs)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateKubernetesConfig(t *testing.T) {
	testcases := []struct {
		name             string
		success          bool
		kubernetesConfig *workspace.KubernetesConfig
	}{
		{
			name:             "valid kubernetes config",
			success:          true,
			kubernetesConfig: mockValidKubernetesConfig(),
		},
		{
			name:             "invalid kubernetes config empty kubeconfig",
			success:          false,
			kubernetesConfig: &workspace.KubernetesConfig{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateKubernetesConfig(tc.kubernetesConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateTerraformConfig(t *testing.T) {
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
			err := ValidateTerraformConfig(tc.terraformConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateBackendConfigs(t *testing.T) {
	testcases := []struct {
		name           string
		success        bool
		backendConfigs *workspace.BackendConfigs
	}{
		{
			name:           "valid backend configs",
			success:        true,
			backendConfigs: mockValidBackendConfigs(),
		},
		{
			name:    "invalid backend configs multiple backends",
			success: false,
			backendConfigs: &workspace.BackendConfigs{
				Local: &workspace.LocalFileConfig{},
				Mysql: &workspace.MysqlConfig{
					DBName: "test",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBackendConfigs(tc.backendConfigs)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateMysqlConfig(t *testing.T) {
	invalidPort := -1
	testcases := []struct {
		name        string
		success     bool
		mysqlConfig *workspace.MysqlConfig
	}{
		{
			name:        "valid mysql config",
			success:     true,
			mysqlConfig: mockValidMysqlConfig(),
		},
		{
			name:    "invalid mysql config empty dbName",
			success: false,
			mysqlConfig: &workspace.MysqlConfig{
				DBName: "",
				User:   "kusion",
				Host:   "127.0.0.1",
			},
		},
		{
			name:    "invalid mysql config empty user",
			success: false,
			mysqlConfig: &workspace.MysqlConfig{
				DBName: "kusion_db",
				User:   "",
				Host:   "127.0.0.1",
			},
		},
		{
			name:    "invalid mysql config empty host",
			success: false,
			mysqlConfig: &workspace.MysqlConfig{
				DBName: "kusion_db",
				User:   "kusion",
				Host:   "",
			},
		},
		{
			name:    "invalid mysql config invalid port",
			success: false,
			mysqlConfig: &workspace.MysqlConfig{
				DBName: "kusion_db",
				User:   "kusion",
				Host:   "127.0.0.1",
				Port:   &invalidPort,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMysqlConfig(tc.mysqlConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateValidateGenericObjectStorageConfig(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		config  *workspace.GenericObjectStorageConfig
	}{
		{
			name:    "valid generic object storage config",
			success: true,
			config:  mockValidGenericObjectStorageConfig(),
		},
		{
			name:    "invalid generic object storage config empty bucket",
			success: false,
			config:  &workspace.GenericObjectStorageConfig{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGenericObjectStorageConfig(tc.config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateWholeOssConfig(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		ossConfig *workspace.OssConfig
	}{
		{
			name:      "valid oss config",
			success:   true,
			ossConfig: mockValidCompletedOssConfig(),
		},
		{
			name:    "invalid oss config empty endpoint",
			success: false,
			ossConfig: &workspace.OssConfig{
				GenericObjectStorageConfig: workspace.GenericObjectStorageConfig{
					Endpoint:        "",
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion_bucket",
				},
			},
		},
		{
			name:    "invalid oss config empty access key id",
			success: false,
			ossConfig: &workspace.OssConfig{
				GenericObjectStorageConfig: workspace.GenericObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					AccessKeyID:     "",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion_bucket",
				},
			},
		},
		{
			name:    "invalid oss config empty access key secret",
			success: false,
			ossConfig: &workspace.OssConfig{
				GenericObjectStorageConfig: workspace.GenericObjectStorageConfig{
					Endpoint:        "http://oss-cn-hangzhou.aliyuncs.com",
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "",
					Bucket:          "kusion_bucket",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWholeOssConfig(tc.ossConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateWholeS3Config(t *testing.T) {
	testcases := []struct {
		name     string
		success  bool
		s3Config *workspace.S3Config
	}{
		{
			name:     "valid s3 config",
			success:  true,
			s3Config: mockValidCompletedS3Config(),
		},
		{
			name:    "invalid s3 config empty region",
			success: false,
			s3Config: &workspace.S3Config{
				GenericObjectStorageConfig: workspace.GenericObjectStorageConfig{
					AccessKeyID:     "fake-access-key-id",
					AccessKeySecret: "fake-access-key-secret",
					Bucket:          "kusion_bucket",
				},
				Region: "",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWholeS3Config(tc.s3Config)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}
