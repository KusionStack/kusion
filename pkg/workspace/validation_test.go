package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockValidWorkspace(name string) *v1.Workspace {
	return &v1.Workspace{
		Name:    name,
		Modules: mockValidModuleConfigs(),
		Context: map[string]any{
			"Kubernetes": map[string]string{
				"Config": "/etc/kubeconfig.yaml",
			},
		},
	}
}

func mockValidModuleConfigs() map[string]*v1.ModuleConfig {
	return map[string]*v1.ModuleConfig{
		"mysql": {
			Path:    "ghcr.io/kusionstack/mysql",
			Version: "0.1.0",
			Configs: v1.Configs{
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
		},
		"network": {
			Path:    "ghcr.io/kusionstack/network",
			Version: "0.1.0",
			Configs: v1.Configs{
				Default: v1.GenericConfig{
					"type": "aws",
				},
			},
		},
	}
}

func mockInvalidModuleConfigs() map[string]v1.ModuleConfig {
	return map[string]v1.ModuleConfig{
		"empty default block": {
			Configs: v1.Configs{
				Default: v1.GenericConfig{},
			},
		},
		"not empty projectSelector in default block": {
			Configs: v1.Configs{
				Default: v1.GenericConfig{
					"type":            "aws",
					"version":         "5.7",
					"instanceType":    "db.t3.micro",
					"projectSelector": []string{"foo", "bar"},
				},
			},
		},
		"empty patcher block name": {
			Configs: v1.Configs{
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
		},
		"empty patcher block": {
			Configs: v1.Configs{
				Default: v1.GenericConfig{
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.micro",
				},
				ModulePatcherConfigs: v1.ModulePatcherConfigs{
					"smallClass": nil,
				},
			},
		},
		"empty config in patcher block": {
			Configs: v1.Configs{
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
		},
		"empty project selector in patcher block": {
			Configs: v1.Configs{
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
		},
		"empty project name": {
			Configs: v1.Configs{
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
		},
		"repeated projects in one patcher block": {
			Configs: v1.Configs{
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
		},
		"repeated projects in multiple patcher blocks": {
			Configs: v1.Configs{
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
			moduleConfig: *mockValidModuleConfigs()["mysql"],
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
			err := ValidateModuleConfig("mysql", &tc.moduleConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestValidateModuleMetadata(t *testing.T) {
	t.Run("ValidModuleMetadata", func(t *testing.T) {
		err := ValidateModuleMetadata("testModule", &v1.ModuleConfig{Version: "1.0.0", Path: "/path/to/module"})
		assert.NoError(t, err)
	})

	t.Run("IgnoreModule", func(t *testing.T) {
		err := ValidateModuleMetadata("service", &v1.ModuleConfig{Version: "1.0.0", Path: "/path/to/module"})
		assert.NoError(t, err)
	})

	t.Run("EmptyModuleVersion", func(t *testing.T) {
		err := ValidateModuleMetadata("testModule", &v1.ModuleConfig{Version: "", Path: "/path/to/module"})
		assert.Error(t, err)
	})

	t.Run("EmptyModulePath", func(t *testing.T) {
		err := ValidateModuleMetadata("testModule", &v1.ModuleConfig{Version: "1.0.0", Path: ""})
		assert.Error(t, err)
	})
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
		spec *v1.SecretStore
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "missing provider spec",
			args: args{
				spec: &v1.SecretStore{},
			},
			want: []error{ErrMissingProvider},
		},
		{
			name: "missing provider type",
			args: args{
				spec: &v1.SecretStore{
					Provider: &v1.ProviderSpec{},
				},
			},
			want: []error{ErrMissingProviderType},
		},
		{
			name: "multi secret store providers",
			args: args{
				spec: &v1.SecretStore{
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
				spec: &v1.SecretStore{
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
