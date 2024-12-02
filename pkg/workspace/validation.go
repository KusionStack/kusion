package workspace

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

var (
	ErrEmptyWorkspaceName = errors.New("empty workspace name")

	ErrEmptyModuleName                      = errors.New("empty module name")
	ErrEmptyModuleConfig                    = errors.New("empty module config")
	ErrEmptyModuleConfigBlock               = errors.New("empty config of a module block")
	ErrEmptyModuleConfigPatcherBlock        = errors.New("empty patcher block in module config")
	ErrEmptyModuleConfigPatcherBlockName    = errors.New("empty patcher block name in module config")
	ErrInvalidModuleConfigPatcherBlockName  = errors.New("patcher name must not be default in module config")
	ErrEmptyModuleConfigProjectSelector     = errors.New("empty projectSelector in module config patcher block")
	ErrNotEmptyModuleConfigProjectSelector  = errors.New("not empty projectSelector in module config default block")
	ErrEmptyModuleConfigProjectName         = errors.New("empty project name at projectSelector in module config patcher block")
	ErrRepeatedModuleConfigSelectedProjects = errors.New("project should not repeat in one patcher block's projectSelector")
	ErrMultipleModuleConfigSelectedProjects = errors.New("a project cannot assign in more than one patcher block's projectSelector")
	ErrMissingProvider                      = errors.New("invalid secret store spec, missing provider config")
	ErrMultiSecretStoreProviders            = errors.New("may not specify more than 1 secret store provider")
	ErrEmptyAWSRegion                       = errors.New("region must be provided when using AWS Secrets Manager")
	ErrEmptyVaultServer                     = errors.New("server address must be provided when using Hashicorp Vault")
	ErrEmptyVaultURL                        = errors.New("vault url must be provided when using Azure KeyVault")
	ErrEmptyTenantID                        = errors.New("azure tenant id must be provided when using Azure KeyVault")
	ErrEmptyAlicloudRegion                  = errors.New("region must be provided when using Alicloud Secrets Manager")
	ErrMissingProviderType                  = errors.New("must specify a provider type")
	ErrInvalidViettelCloudProjectID         = errors.New("invalid format project id for ViettelCloud Secrets Manager")
)

// ValidateWorkspace is used to validate the workspace get or set in the storage.
func ValidateWorkspace(ws *v1.Workspace) error {
	if ws.Name == "" {
		return ErrEmptyWorkspaceName
	}
	if ws.Modules != nil {
		if err := ValidateModuleConfigs(ws.Modules); err != nil {
			return err
		}
	}
	if ws.SecretStore != nil {
		if allErrs := ValidateSecretStoreConfig(ws.SecretStore); allErrs != nil {
			return utilerrors.NewAggregate(allErrs)
		}
	}
	return nil
}

// ValidateModuleConfigs validates the moduleConfigs is valid or not.
func ValidateModuleConfigs(configs v1.ModuleConfigs) error {
	for name, cfg := range configs {
		if name == "" {
			return ErrEmptyModuleName
		}
		if cfg == nil {
			return fmt.Errorf("%w, module name: %s", ErrEmptyModuleConfig, name)
		}
		if err := ValidateModuleConfig(name, cfg); err != nil {
			return fmt.Errorf("%w, module name: %s", err, name)
		}
	}

	return nil
}

// ValidateModuleConfig is used to validate the moduleConfig is valid or not.
func ValidateModuleConfig(name string, config *v1.ModuleConfig) error {
	if err := ValidateModuleMetadata(name, config); err != nil {
		return err
	}
	if err := ValidateModuleDefaultConfig(config.Configs.Default); err != nil {
		return err
	}
	if err := ValidateModulePatcherConfigs(config.Configs.ModulePatcherConfigs); err != nil {
		return err
	}
	return nil
}

func ValidateModuleMetadata(name string, config *v1.ModuleConfig) error {
	if config.Version == "" {
		return fmt.Errorf("empty version of module:%s in the workspacek config", name)
	}
	if config.Path == "" {
		return fmt.Errorf("empty path of module:%s in the workspacek config", name)
	}
	return nil
}

func ValidateModuleDefaultConfig(config v1.GenericConfig) error {
	if config == nil {
		return nil
	}
	if _, ok := config[v1.ProjectSelectorField]; ok {
		return ErrNotEmptyModuleConfigProjectSelector
	}
	return nil
}

func ValidateModulePatcherConfigs(config v1.ModulePatcherConfigs) error {
	// allProjects is used to inspect if there are repeated projects in projectSelector
	// field or not.
	allProjects := make(map[string]string)
	for name, cfg := range config {
		switch name {
		case "":
			return ErrEmptyModuleConfigPatcherBlockName

		// name of patcher block must not be default.
		case v1.DefaultBlock:
			return ErrInvalidModuleConfigPatcherBlockName

		// repeated projects in different patcher blocks are not allowed.
		default:
			if cfg == nil {
				return fmt.Errorf("%w, patcher block: %s", ErrEmptyModuleConfigPatcherBlock, name)
			}
			if len(cfg.GenericConfig) == 0 {
				return fmt.Errorf("%w, patcher block: %s", ErrEmptyModuleConfigBlock, name)
			}
			if len(cfg.ProjectSelector) == 0 {
				return fmt.Errorf("%w, patcher block: %s", ErrEmptyModuleConfigProjectSelector, name)
			}

			// a project cannot assign in more than one patcher block.
			for _, project := range cfg.ProjectSelector {
				if project == "" {
					return fmt.Errorf("%w, patcher block: %s", ErrEmptyModuleConfigProjectName, name)
				}

				patcherName, ok := allProjects[project]
				if ok {
					if patcherName == name {
						return fmt.Errorf("%w, patcher block: %s", ErrRepeatedModuleConfigSelectedProjects, name)
					} else {
						return fmt.Errorf("%w, conflict patcher block: %s, %s", ErrMultipleModuleConfigSelectedProjects, name, patcherName)
					}
				}
				allProjects[project] = name
			}
		}
	}

	return nil
}

// ValidateSecretStoreConfig tests that the specified SecretStore has valid data.
func ValidateSecretStoreConfig(spec *v1.SecretStore) []error {
	if spec.Provider == nil {
		return []error{ErrMissingProvider}
	}

	numProviders := 0
	var allErrs []error
	if spec.Provider.AWS != nil {
		numProviders++
		allErrs = append(allErrs, validateAWSSecretStore(spec.Provider.AWS)...)
	}
	if spec.Provider.Vault != nil {
		if numProviders > 0 {
			allErrs = append(allErrs, ErrMultiSecretStoreProviders)
		} else {
			numProviders++
			allErrs = append(allErrs, validateHashiVaultSecretStore(spec.Provider.Vault)...)
		}
	}
	if spec.Provider.Azure != nil {
		if numProviders > 0 {
			allErrs = append(allErrs, ErrMultiSecretStoreProviders)
		} else {
			numProviders++
			allErrs = append(allErrs, validateAzureKeyVaultSecretStore(spec.Provider.Azure)...)
		}
	}
	if spec.Provider.Alicloud != nil {
		if numProviders > 0 {
			allErrs = append(allErrs, ErrMultiSecretStoreProviders)
		} else {
			numProviders++
			allErrs = append(allErrs, validateAlicloudSecretStore(spec.Provider.Alicloud)...)
		}
	}

	if spec.Provider.ViettelCloud != nil {
		if numProviders > 0 {
			allErrs = append(allErrs, ErrMultiSecretStoreProviders)
		} else {
			numProviders++
			allErrs = append(allErrs, validateViettelCloudSecretStore(spec.Provider.ViettelCloud)...)
		}
	}

	if numProviders == 0 {
		allErrs = append(allErrs, ErrMissingProviderType)
	}

	return allErrs
}

func validateAWSSecretStore(ss *v1.AWSProvider) []error {
	var allErrs []error
	if len(ss.Region) == 0 {
		allErrs = append(allErrs, ErrEmptyAWSRegion)
	}
	return allErrs
}

func validateHashiVaultSecretStore(vault *v1.VaultProvider) []error {
	var allErrs []error
	if len(vault.Server) == 0 {
		allErrs = append(allErrs, ErrEmptyVaultServer)
	}
	return allErrs
}

func validateAzureKeyVaultSecretStore(azureKv *v1.AzureKVProvider) []error {
	var allErrs []error
	if azureKv.VaultURL == nil || len(*azureKv.VaultURL) == 0 {
		allErrs = append(allErrs, ErrEmptyVaultURL)
	}
	if azureKv.TenantID == nil || len(*azureKv.TenantID) == 0 {
		allErrs = append(allErrs, ErrEmptyTenantID)
	}
	return allErrs
}

func validateAlicloudSecretStore(ac *v1.AlicloudProvider) []error {
	var allErrs []error
	if len(ac.Region) == 0 {
		allErrs = append(allErrs, ErrEmptyAlicloudRegion)
	}
	return allErrs
}

func validateViettelCloudSecretStore(vc *v1.ViettelCloudProvider) []error {
	var allErrs []error
	if vc.ProjectID != "" {
		if _, err := uuid.Parse(vc.ProjectID); err != nil {
			allErrs = append(allErrs, ErrInvalidViettelCloudProjectID)
		}
	}
	return allErrs
}
