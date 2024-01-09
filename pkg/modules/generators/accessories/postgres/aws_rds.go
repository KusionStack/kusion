package postgres

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
)

const (
	defaultAWSProviderURL = "registry.terraform.io/hashicorp/aws/5.0.1"
	awsRegionEnv          = "AWS_REGION"
	awsSecurityGroup      = "aws_security_group"
	awsDBInstance         = "aws_db_instance"
)

type awsSecurityGroupTraffic struct {
	CidrBlocks     []string `yaml:"cidr_blocks" json:"cidr_blocks"`
	Description    string   `yaml:"description" json:"description"`
	FromPort       int      `yaml:"from_port" json:"from_port"`
	IPv6CIDRBlocks []string `yaml:"ipv6_cidr_blocks" json:"ipv6_cidr_blocks"`
	PrefixListIDs  []string `yaml:"prefix_list_ids" json:"prefix_list_ids"`
	Protocol       string   `yaml:"protocol" json:"protocol"`
	SecurityGroups []string `yaml:"security_groups" json:"security_groups"`
	Self           bool     `yaml:"self" json:"self"`
	ToPort         int      `yaml:"to_port" json:"to_port"`
}

// generateAWSResources generates aws provided postgresql database instance.
func (g *postgresGenerator) generateAWSResources(db *postgres.PostgreSQL, spec *apiv1.Intent) (*v1.Secret, error) {
	// Set the terraform random and aws provider.
	randomProvider, awsProvider := &inputs.Provider{}, &inputs.Provider{}

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

	awsProviderCfg, ok := g.tfConfigs[inputs.AWSProvider]
	if !ok {
		awsProvider.SetString(defaultAWSProviderURL)
	} else {
		awsProviderURL, err := inputs.GetProviderURL(awsProviderCfg)
		if err != nil {
			return nil, err
		}
		if err := awsProvider.SetString(awsProviderURL); err != nil {
			return nil, err
		}
	}

	// Get the aws provider region, and the region of the aws provider must be set.
	var awsProviderRegion string
	if awsProviderRegion = inputs.GetProviderRegion(g.tfConfigs[inputs.AWSProvider]); awsProviderRegion == "" {
		awsProviderRegion = os.Getenv(awsRegionEnv)
	}
	if awsProviderRegion == "" {
		return nil, fmt.Errorf("aws provider region should not be empty")
	}

	// Build random_password for aws_db_instance.
	randomPasswordID, r := g.generateTFRandomPassword(randomProvider)
	spec.Resources = append(spec.Resources, r)

	// Build aws_security group for aws_db_instance.
	awsSecurityGroupID, r, err := g.generateAWSSecurityGroup(awsProvider, awsProviderRegion, db)
	if err != nil {
		return nil, err
	}
	spec.Resources = append(spec.Resources, r)

	// Build aws_db_instance.
	awsDBInstanceID, r := g.generateAWSDBInstance(awsProviderRegion, awsSecurityGroupID, randomPasswordID, awsProvider, db)
	spec.Resources = append(spec.Resources, r)

	// Inject the database host address, username and password into k8s secret.
	hostAddress := modules.KusionPathDependency(awsDBInstanceID, "address")
	password := modules.KusionPathDependency(randomPasswordID, "result")

	return g.generateDBSecret(hostAddress, db.Username, password, spec)
}

func (g *postgresGenerator) generateAWSSecurityGroup(
	provider *inputs.Provider,
	region string,
	db *postgres.PostgreSQL,
) (string, apiv1.Resource, error) {
	// SecurityIPs should be in the format of IP address or Classes Inter-Domain
	// Routing (CIDR) mode.
	for _, ip := range db.SecurityIPs {
		if !isIPAddress(ip) && !isCIDR(ip) {
			return "", apiv1.Resource{}, fmt.Errorf("illegal security ip format: %v", ip)
		}
	}

	sgAttrs := map[string]interface{}{
		"egress": []awsSecurityGroupTraffic{
			{
				CidrBlocks: []string{"0.0.0.0/0"},
				Protocol:   "-1",
				FromPort:   0,
				ToPort:     0,
			},
		},
		"ingress": []awsSecurityGroupTraffic{
			{
				CidrBlocks: db.SecurityIPs,
				Protocol:   "tcp",
				FromPort:   5432,
				ToPort:     5432,
			},
		},
	}

	id := modules.TerraformResourceID(provider, awsSecurityGroup, g.postgres.DatabaseName+dbResSuffix)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, awsSecurityGroup)

	return id, modules.TerraformResource(id, nil, sgAttrs, pvdExts), nil
}

func (g *postgresGenerator) generateAWSDBInstance(
	region, awsSecurityGroupID, randomPasswordID string,
	provider *inputs.Provider, db *postgres.PostgreSQL,
) (string, apiv1.Resource) {
	dbAttrs := map[string]interface{}{
		"allocated_storage":   db.Size,
		"engine":              dbEngine,
		"engine_version":      db.Version,
		"identifier":          db.DatabaseName,
		"instance_class":      db.InstanceType,
		"password":            modules.KusionPathDependency(randomPasswordID, "result"),
		"publicly_accessible": isPublicAccessible(db.SecurityIPs),
		"skip_final_snapshot": true,
		"username":            db.Username,
		"vpc_security_group_ids": []string{
			modules.KusionPathDependency(awsSecurityGroupID, "id"),
		},
	}

	if db.SubnetID != "" {
		dbAttrs["db_subnet_group_name"] = db.SubnetID
	}

	id := modules.TerraformResourceID(provider, awsDBInstance, db.DatabaseName)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, awsDBInstance)

	return id, modules.TerraformResource(id, nil, dbAttrs, pvdExts)
}
