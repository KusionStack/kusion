package accessories

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/database"
)

const (
	awsSecurityGroup   = "aws_security_group"
	awsDBInstance      = "aws_db_instance"
	defaultAWSProvider = "registry.terraform.io/hashicorp/aws/5.0.1"
)

var (
	tfProviderAWS     = os.Getenv("TF_PROVIDER_AWS")
	awsProviderRegion = os.Getenv("AWS_PROVIDER_REGION")
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

func (g *databaseGenerator) generateAWSResources(db *database.Database, spec *intent.Intent) (*v1.Secret, error) {
	// Set the terraform random and aws provider.
	randomProvider := &inputs.Provider{}
	if err := randomProvider.SetString(randomProviderURL); err != nil {
		return nil, err
	}

	// The region of the aws provider must be set.
	if awsProviderRegion == "" {
		return nil, fmt.Errorf("the region of the aws provider must be set")
	}

	var providerURL string
	awsProvider := &inputs.Provider{}
	if tfProviderAWS == "" {
		providerURL = defaultAWSProvider
	} else {
		providerURL = tfProviderAWS
	}

	if err := awsProvider.SetString(providerURL); err != nil {
		return nil, err
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

func (g *databaseGenerator) generateAWSSecurityGroup(
	provider *inputs.Provider,
	region string,
	db *database.Database,
) (string, intent.Resource, error) {
	// SecurityIPs should be in the format of IP address or Classes Inter-Domain
	// Routing (CIDR) mode.
	for _, ip := range db.SecurityIPs {
		if !isIPAddress(ip) && !isCIDR(ip) {
			return "", intent.Resource{}, fmt.Errorf("illegal security ip format: %v", ip)
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
				FromPort:   3306,
				ToPort:     3306,
			},
		},
	}

	id := modules.TerraformResourceID(provider, awsSecurityGroup, g.appName+dbResSuffix)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, awsSecurityGroup)

	return id, modules.TerraformResource(id, nil, sgAttrs, pvdExts), nil
}

func (g *databaseGenerator) generateAWSDBInstance(
	region, awsSecurityGroupID, randomPasswordID string,
	provider *inputs.Provider, db *database.Database,
) (string, intent.Resource) {
	dbAttrs := map[string]interface{}{
		"allocated_storage":   db.Size,
		"engine":              db.Engine,
		"engine_version":      db.Version,
		"identifier":          g.appName,
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

	id := modules.TerraformResourceID(provider, awsDBInstance, g.appName)
	pvdExts := modules.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, awsDBInstance)

	return id, modules.TerraformResource(id, nil, dbAttrs, pvdExts)
}
