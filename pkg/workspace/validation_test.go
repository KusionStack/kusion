package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockValidWorkspace(name string) *v1.Workspace {
	return &v1.Workspace{
		Name:     name,
		Modules:  mockValidModuleConfigs(),
		Runtimes: mockValidRuntimeConfigs(),
	}
}

func mockValidModuleConfigs() map[string]*v1.ModuleConfig {
	return map[string]*v1.ModuleConfig{
		"database": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "bar"},
				},
			},
		},
		"port": {
			Default: v1.GenericConfig{
				"type": "aws",
			},
		},
	}
}

func mockInvalidModuleConfigs() map[string]v1.ModuleConfig {
	return map[string]v1.ModuleConfig{
		"empty default block": {
			Default: v1.GenericConfig{},
		},
		"not empty projectSelector in default block": {
			Default: v1.GenericConfig{
				"type":            "aws",
				"version":         "5.7",
				"instanceType":    "db.t3.micro",
				"projectSelector": []string{"foo", "bar"},
			},
		},
		"empty patcher block name": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "bar"},
				},
			},
		},
		"empty patcher block": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": nil,
			},
		},
		"empty config in patcher block": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": {
					ProjectSelector: []string{"foo", "bar"},
				},
			},
		},
		"empty project selector in patcher block": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.small",
					},
				},
			},
		},
		"empty project name": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"", "bar"},
				},
			},
		},
		"repeated projects in one patcher block": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "foo"},
				},
			},
		},
		"repeated projects in multiple patcher blocks": {
			Default: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
			ModulePatcherConfigs: v1.ModulePatcherConfigs{
				"smallClass": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.small",
					},
					ProjectSelector: []string{"foo", "bar"},
				},
				"largeClass": {
					GenericConfig: v1.GenericConfig{
						"instanceType": "db.t3.large",
					},
					ProjectSelector: []string{"foo"},
				},
			},
		},
	}
}

func mockValidRuntimeConfigs() *v1.RuntimeConfigs {
	return &v1.RuntimeConfigs{
		Kubernetes: mockValidKubernetesConfig(),
		Terraform:  mockValidTerraformConfig(),
	}
}

func mockValidKubernetesConfig() *v1.KubernetesConfig {
	return &v1.KubernetesConfig{
		KubeConfig: "/etc/kubeconfig.yaml",
	}
}

func mockValidTerraformConfig() v1.TerraformConfig {
	return v1.TerraformConfig{
		"aws": {
			Source:  "hashicorp/aws",
			Version: "1.0.4",
			GenericConfig: v1.GenericConfig{
				"region": "us-east-1",
			},
		},
	}
}

func TestValidateWorkspace(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		workspace *v1.Workspace
	}{
		{
			name:      "valid workspace",
			success:   true,
			workspace: mockValidWorkspace("dev"),
		},
		{
			name:      "invalid workspace empty name",
			success:   false,
			workspace: &v1.Workspace{},
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
		moduleConfigs v1.ModuleConfigs
	}{
		{
			name:          "valid module configs",
			success:       true,
			moduleConfigs: mockValidModuleConfigs(),
		},
		{
			name:    "invalid module configs empty module name",
			success: false,
			moduleConfigs: v1.ModuleConfigs{
				"": mockValidModuleConfigs()["database"],
			},
		},
		{
			name:    "invalid module configs empty module config",
			success: false,
			moduleConfigs: v1.ModuleConfigs{
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
		moduleConfig v1.ModuleConfig
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
			moduleConfig v1.ModuleConfig
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
		runtimeConfigs *v1.RuntimeConfigs
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
		kubernetesConfig *v1.KubernetesConfig
	}{
		{
			name:             "valid kubernetes config",
			success:          true,
			kubernetesConfig: mockValidKubernetesConfig(),
		},
		{
			name:             "invalid kubernetes config empty kubeconfig",
			success:          false,
			kubernetesConfig: &v1.KubernetesConfig{},
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
		terraformConfig v1.TerraformConfig
	}{
		{
			name:            "valid terraform config",
			success:         true,
			terraformConfig: mockValidTerraformConfig(),
		},
		{
			name:    "invalid terraform config empty provider name",
			success: false,
			terraformConfig: v1.TerraformConfig{
				"": {
					Source:  "hashicorp/aws",
					Version: "1.0.4",
					GenericConfig: v1.GenericConfig{
						"region": "us-east-1",
					},
				},
			},
		},
		{
			name:    "invalid terraform config empty provider config",
			success: false,
			terraformConfig: v1.TerraformConfig{
				"aws": nil,
			},
		},
		{
			name:    "invalid terraform config empty provider source",
			success: false,
			terraformConfig: v1.TerraformConfig{
				"aws": {
					Source:  "",
					Version: "1.0.4",
					GenericConfig: v1.GenericConfig{
						"region": "us-east-1",
					},
				},
			},
		},
		{
			name:    "invalid terraform config empty provider version",
			success: false,
			terraformConfig: v1.TerraformConfig{
				"aws": {
					Source:  "hashicorp/aws",
					Version: "",
					GenericConfig: v1.GenericConfig{
						"region": "us-east-1",
					},
				},
			},
		},
		{
			name:    "invalid terraform config empty provider config key",
			success: false,
			terraformConfig: v1.TerraformConfig{
				"aws": {
					Source:  "hashicorp/aws",
					Version: "1.0.4",
					GenericConfig: v1.GenericConfig{
						"": "us-east-1",
					},
				},
			},
		},
		{
			name:    "invalid terraform config empty provider config value",
			success: false,
			terraformConfig: v1.TerraformConfig{
				"aws": {
					Source:  "hashicorp/aws",
					Version: "1.0.4",
					GenericConfig: v1.GenericConfig{
						"region": nil,
					},
				},
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

func TestValidateAWSSecretStore(t *testing.T) {
	type args struct {
		ss *v1.AWSProvider
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "valid AWS provider spec",
			args: args{
				ss: &v1.AWSProvider{
					Region: "eu-west-2",
				},
			},
			want: nil,
		},
		{
			name: "invalid AWS provider spec",
			args: args{
				ss: &v1.AWSProvider{},
			},
			want: []error{ErrEmptyAWSRegion},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateAWSSecretStore(tt.args.ss), "validateAWSSecretStore(%v)", tt.args.ss)
		})
	}
}

func TestValidateHashiVaultSecretStore(t *testing.T) {
	type args struct {
		vault *v1.VaultProvider
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "valid Hashi Vault provider spec",
			args: args{
				vault: &v1.VaultProvider{
					Server: "https://vault.example.com:8200",
				},
			},
			want: nil,
		},
		{
			name: "invalid Hashi Vault provider spec",
			args: args{
				vault: &v1.VaultProvider{},
			},
			want: []error{ErrEmptyVaultServer},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateHashiVaultSecretStore(tt.args.vault), "validateHashiVaultSecretStore(%v)", tt.args.vault)
		})
	}
}

func TestValidateAzureKeyVaultSecretStore(t *testing.T) {
	type args struct {
		azureKv *v1.AzureKVProvider
	}
	vaultURL := "https://local.vault.url"
	tenantID := "my-tenant-id"
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "valid Azure KV provider spec",
			args: args{
				azureKv: &v1.AzureKVProvider{
					VaultURL: &vaultURL,
					TenantID: &tenantID,
				},
			},
			want: nil,
		},
		{
			name: "invalid Azure KV provider spec",
			args: args{
				azureKv: &v1.AzureKVProvider{},
			},
			want: []error{ErrEmptyVaultURL, ErrEmptyTenantID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateAzureKeyVaultSecretStore(tt.args.azureKv), "validateAzureKeyVaultSecretStore(%v)", tt.args.azureKv)
		})
	}
}

func TestValidateAlicloudSecretStore(t *testing.T) {
	type args struct {
		ac *v1.AlicloudProvider
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "valid Alicloud provider spec",
			args: args{
				ac: &v1.AlicloudProvider{
					Region: "sh",
				},
			},
			want: nil,
		},
		{
			name: "invalid Alicloud provider spec",
			args: args{
				ac: &v1.AlicloudProvider{},
			},
			want: []error{ErrEmptyAlicloudRegion},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateAlicloudSecretStore(tt.args.ac), "validateAlicloudSecretStore(%v)", tt.args.ac)
		})
	}
}

func TestValidateSecretStoreConfig(t *testing.T) {
	type args struct {
		spec *v1.SecretStoreSpec
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "missing provider spec",
			args: args{
				spec: &v1.SecretStoreSpec{},
			},
			want: []error{ErrMissingProvider},
		},
		{
			name: "missing provider type",
			args: args{
				spec: &v1.SecretStoreSpec{
					Provider: &v1.ProviderSpec{},
				},
			},
			want: []error{ErrMissingProviderType},
		},
		{
			name: "multi secret store providers",
			args: args{
				spec: &v1.SecretStoreSpec{
					Provider: &v1.ProviderSpec{
						AWS: &v1.AWSProvider{
							Region: "us-east-1",
						},
						Vault: &v1.VaultProvider{
							Server: "https://vault.example.com:8200",
						},
					},
				},
			},
			want: []error{ErrMultiSecretStoreProviders},
		},
		{
			name: "valid secret store spec",
			args: args{
				spec: &v1.SecretStoreSpec{
					Provider: &v1.ProviderSpec{
						AWS: &v1.AWSProvider{
							Region: "us-east-1",
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ValidateSecretStoreConfig(tt.args.spec), "validateAlicloudSecretStore(%v)", tt.args.spec)
		})
	}
}
