package postgres

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
)

const (
	defaultAlicloudProviderURL = "registry.terraform.io/aliyun/alicloud/1.209.1"
	alicloudRegionEnv          = "ALICLOUD_REGION"
	alicloudDBInstance         = "alicloud_db_instance"
	alicloudDBConnection       = "alicloud_db_connection"
	alicloudRDSAccount         = "alicloud_rds_account"
)

type alicloudServerlessConfig struct {
	AutoPause   bool `yaml:"auto_pause" json:"auto_pause"`
	SwitchForce bool `yaml:"switch_force" json:"switch_force"`
	MaxCapacity int  `yaml:"max_capacity,omitempty" json:"max_capacity,omitempty"`
	MinCapacity int  `yaml:"min_capacity,omitempty" json:"min_capacity,omitempty"`
}

// generateAlicloudResources generates alicloud provided postgresql database instance.
func (g *postgresGenerator) generateAlicloudResources(db *postgres.PostgreSQL, spec *apiv1.Intent) (*v1.Secret, error) {
	// Set the terraform random and alicloud provider.
	randomProvider, alicloudProvider := &inputs.Provider{}, &inputs.Provider{}

	randomProviderCfg, ok := g.tfConfigs[inputs.RandomProvider]
	if !ok {
		randomProvider.SetString(defaultRandomProviderURL)
	} else {
		randomProviderURL, err := inputs.GetProviderURL(randomProviderCfg)
		if err != nil {
			return nil, err
		}
		if err := randomProvider.SetString(randomProviderURL); err != nil {
			return nil, err
		}
	}

	alicloudProviderCfg, ok := g.tfConfigs[inputs.AlicloudProvider]
	if !ok {
		alicloudProvider.SetString(defaultAlicloudProviderURL)
	} else {
		alicloudProviderURL, err := inputs.GetProviderURL(alicloudProviderCfg)
		if err != nil {
			return nil, err
		}
		if err := alicloudProvider.SetString(alicloudProviderURL); err != nil {
			return nil, err
		}
	}

	// Get the alicloud provider region, and the region of the alicloud provider must be set.
	var alicloudProviderRegion string
	if alicloudProviderRegion = inputs.GetProviderRegion(g.tfConfigs[inputs.AlicloudProvider]); alicloudProviderRegion == "" {
		alicloudProviderRegion = os.Getenv(alicloudRegionEnv)
	}
	if alicloudProviderRegion == "" {
		return nil, fmt.Errorf("alicloud provider region should not be empty")
	}

	// Build alicloud_db_instance.
	alicloudDBInstanceID, r := g.generateAlicloudDBInstance(alicloudProviderRegion, alicloudProvider, db)
	spec.Resources = append(spec.Resources, r)

	// Build alicloud_db_connection for alicloud_db_instance.
	var alicloudDBConnectionID string
	if isPublicAccessible(db.SecurityIPs) {
		alicloudDBConnectionID, r = g.generateAlicloudDBConnection(alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider)
		spec.Resources = append(spec.Resources, r)
	}

	// Build random_password for alicloud_rds_account.
	randomPasswordID, r := g.generateTFRandomPassword(randomProvider)
	spec.Resources = append(spec.Resources, r)

	// Build alicloud_rds_account.
	r = g.generateAlicloudRDSAccount(db.Username, randomPasswordID, alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider, db)
	spec.Resources = append(spec.Resources, r)

	// Inject the host address, username and password into k8s secret.
	password := modules.KusionPathDependency(randomPasswordID, "result")
	hostAddress := modules.KusionPathDependency(alicloudDBInstanceID, "connection_string")
	if !db.PrivateRouting {
		// Set the public network connection string as the host address.
		hostAddress = modules.KusionPathDependency(alicloudDBConnectionID, "connection_string")
	}

	return g.generateDBSecret(hostAddress, db.Username, password, spec)
}

func (g *postgresGenerator) generateAlicloudDBInstance(
	region string,
	provider *inputs.Provider,
	db *postgres.PostgreSQL,
) (string, apiv1.Resource) {
	dbAttrs := map[string]interface{}{
		"category":         db.Category,
		"engine":           "PostgreSQL",
		"engine_version":   db.Version,
		"instance_storage": db.Size,
		"instance_type":    db.InstanceType,
		"security_ips":     db.SecurityIPs,
		"vswitch_id":       db.SubnetID,
		"instance_name":    db.DatabaseName,
	}

	// Set serverless specific attributes.
	if strings.Contains(db.Category, "serverless") {
		dbAttrs["db_instance_storage_type"] = "cloud_essd"
		dbAttrs["instance_charge_type"] = "Serverless"

		serverlessConfig := alicloudServerlessConfig{
			MaxCapacity: 12,
			MinCapacity: 1,
		}
		serverlessConfig.AutoPause = false
		serverlessConfig.SwitchForce = false

		dbAttrs["serverless_config"] = []alicloudServerlessConfig{
			serverlessConfig,
		}
	}

	id := modules.TerraformResourceID(provider, alicloudDBInstance, db.DatabaseName)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, alicloudDBInstance)

	return id, modules.TerraformResource(id, nil, dbAttrs, pvdExts)
}

func (g *postgresGenerator) generateAlicloudDBConnection(
	dbInstanceID, region string,
	provider *inputs.Provider,
) (string, apiv1.Resource) {
	dbConnectionAttrs := map[string]interface{}{
		"instance_id": modules.KusionPathDependency(dbInstanceID, "id"),
	}

	id := modules.TerraformResourceID(provider, alicloudDBConnection, g.postgres.DatabaseName)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, alicloudDBConnection)

	return id, modules.TerraformResource(id, nil, dbConnectionAttrs, pvdExts)
}

func (g *postgresGenerator) generateAlicloudRDSAccount(
	accountName, randomPasswordID, dbInstanceID, region string,
	provider *inputs.Provider, db *postgres.PostgreSQL,
) apiv1.Resource {
	rdsAccountAttrs := map[string]interface{}{
		"account_name":     accountName,
		"account_password": modules.KusionPathDependency(randomPasswordID, "result"),
		"account_type":     "Super",
		"db_instance_id":   modules.KusionPathDependency(dbInstanceID, "id"),
	}

	id := modules.TerraformResourceID(provider, alicloudRDSAccount, db.DatabaseName)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, alicloudRDSAccount)

	return modules.TerraformResource(id, nil, rdsAccountAttrs, pvdExts)
}
