package accessories

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/accessories/database"
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
	CidrBlocks []string `yaml:"cidr_blocks,omitempty" json:"cidr_blocks,omitempty"`
	Protocol   string   `yaml:"protocol,omitempty" json:"protoco,omitempty"`
	FromPort   int      `yaml:"from_port,omitempty" json:"from_port,omitempty"`
	ToPort     int      `yaml:"to_port,omitempty" json:"to_port,omitempty"`
}

func (g *databaseGenerator) generateAWSResources(db *database.Database, spec *models.Spec) (*v1.Secret, error) {
	// Set the terraform random and aws provider.
	randomProvider := &models.Provider{}
	if err := randomProvider.SetString(randomProviderURL); err != nil {
		return nil, err
	}

	// The region of the aws provider must be set.
	if awsProviderRegion == "" {
		return nil, fmt.Errorf("the region of the aws provider must be set")
	}

	var providerURL string
	awsProvider := &models.Provider{}
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
	hostAddress := appconfiguration.KusionPathDependency(awsDBInstanceID, "address")
	password := appconfiguration.KusionPathDependency(randomPasswordID, "result")

	return g.generateDBSecret(hostAddress, db.Username, password, spec)
}

func (g *databaseGenerator) generateAWSSecurityGroup(
	provider *models.Provider,
	region string,
	db *database.Database,
) (string, models.Resource, error) {
	// SecurityIPs should be in the format of IP address or Classes Inter-Domain
	// Routing (CIDR) mode.
	for _, ip := range db.SecurityIPs {
		if !isIPAddress(ip) && !isCIDR(ip) {
			return "", models.Resource{}, fmt.Errorf("illegal security ip format: %v", ip)
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

	id := appconfiguration.TerraformResourceID(provider, awsSecurityGroup, g.appName+dbResSuffix)
	pvdExts := appconfiguration.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, awsSecurityGroup)

	return id, appconfiguration.TerraformResource(id, nil, sgAttrs, pvdExts), nil
}

func (g *databaseGenerator) generateAWSDBInstance(region, awsSecurityGroupID, randomPasswordID string,
	provider *models.Provider, db *database.Database,
) (string, models.Resource) {
	dbAttrs := map[string]interface{}{
		"allocated_storage":   db.Size,
		"engine":              db.Engine,
		"engine_version":      db.Version,
		"identifier":          g.appName,
		"instance_class":      db.InstanceType,
		"password":            appconfiguration.KusionPathDependency(randomPasswordID, "result"),
		"publicly_accessible": isPublicAccessible(db.SecurityIPs),
		"skip_final_snapshot": true,
		"username":            db.Username,
		"vpc_security_groups_ids": []string{
			appconfiguration.KusionPathDependency(awsSecurityGroupID, "id"),
		},
	}

	if db.SubnetID != "" {
		dbAttrs["db_subnet_group_name"] = db.SubnetID
	}

	id := appconfiguration.TerraformResourceID(provider, awsDBInstance, g.appName)
	pvdExts := appconfiguration.ProviderExtensions(provider, map[string]any{
		"region": region,
	}, awsDBInstance)

	return id, appconfiguration.TerraformResource(id, nil, dbAttrs, pvdExts)
}
