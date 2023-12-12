package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules/inputs"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestMySQLGenerator_GenerateAWSResources(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": &database.Database{
			Header: database.Header{
				Type: "MySQL",
			},
			MySQL: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"mysql": apiv1.GenericConfig{
			"cloud":          "aws",
			"size":           20,
			"instanceType":   "db.t3.micro",
			"privateRouting": false,
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"aws": &apiv1.ProviderConfig{
			Version: "5.0.1",
			Source:  "hashicorp/aws",
			GenericConfig: apiv1.GenericConfig{
				"region": "us-east-1",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "db.t3.micro",
		Category:       defaultCategory,
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		DatabaseName:   "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name        string
		db          *mysql.MySQL
		spec        *apiv1.Intent
		expected    *v1.Secret
		expectedErr error
	}{
		{
			name: "Generate AWS Resources",
			db:   db,
			spec: &apiv1.Intent{},
			expected: &v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: v1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testmysql-mysql",
					Namespace: "testproject",
				},
				StringData: map[string]string{
					"hostAddress": "$kusion_path.hashicorp:aws:aws_db_instance:testmysql.address",
					"username":    "root",
					"password":    "$kusion_path.hashicorp:random:random_password:testmysql-mysql.result",
				},
			},
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*mysqlGenerator).generateAWSResources(test.db, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateAWSSecurityGroup(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": &database.Database{
			Header: database.Header{
				Type: "MySQL",
			},
			MySQL: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"mysql": apiv1.GenericConfig{
			"cloud":          "aws",
			"size":           20,
			"instanceType":   "db.t3.micro",
			"privateRouting": false,
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"aws": &apiv1.ProviderConfig{
			Version: "5.0.1",
			Source:  "hashicorp/aws",
			GenericConfig: apiv1.GenericConfig{
				"region": "us-east-1",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "db.t3.micro",
		Category:       defaultCategory,
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		DatabaseName:   "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name        string
		providerURL string
		region      string
		db          *mysql.MySQL
		expectedID  string
		expectedRes apiv1.Resource
		expectedErr error
	}{
		{
			name:        "Generate AWS Security Group",
			providerURL: "registry.terraform.io/hashicorp/aws/5.0.1",
			region:      "us-east-1",
			db:          db,
			expectedID:  "hashicorp:aws:aws_security_group:testmysql-mysql",
			expectedRes: apiv1.Resource{
				ID:   "hashicorp:aws:aws_security_group:testmysql-mysql",
				Type: "Terraform",
				Attributes: map[string]interface{}{
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
							CidrBlocks: []string{"0.0.0.0/0"},
							Protocol:   "tcp",
							FromPort:   3306,
							ToPort:     3306,
						},
					},
				},
				Extensions: map[string]interface{}{
					"provider": "registry.terraform.io/hashicorp/aws/5.0.1",
					"providerMeta": map[string]interface{}{
						"region": "us-east-1",
					},
					"resourceType": awsSecurityGroup,
				},
			},
		},
	}

	for _, test := range tests {
		var awsProvider = &inputs.Provider{}
		_ = awsProvider.SetString(test.providerURL)
		actualID, actualRes, actualErr := g.(*mysqlGenerator).generateAWSSecurityGroup(
			awsProvider, test.region, test.db,
		)

		if test.expectedErr == nil {
			assert.Equal(t, test.expectedID, actualID)
			assert.Equal(t, test.expectedRes, actualRes)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateAWSDBInstance(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": &database.Database{
			Header: database.Header{
				Type: "MySQL",
			},
			MySQL: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"mysql": apiv1.GenericConfig{
			"cloud":          "aws",
			"size":           20,
			"instanceType":   "db.t3.micro",
			"privateRouting": false,
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"aws": &apiv1.ProviderConfig{
			Version: "5.0.1",
			Source:  "hashicorp/aws",
			GenericConfig: apiv1.GenericConfig{
				"region": "us-east-1",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "db.t3.micro",
		Category:       defaultCategory,
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		DatabaseName:   "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name               string
		region             string
		awsSecurityGroupID string
		randomPasswordID   string
		providerURL        string
		db                 *mysql.MySQL
		expectedID         string
		expectedRes        apiv1.Resource
	}{
		{
			name:               "Generate AWS DB Instance",
			region:             "us-east-1",
			awsSecurityGroupID: "hashicorp:aws:aws_security_group:testmysql-mysql",
			randomPasswordID:   "hashicorp:random:random_password:testmysql-mysql",
			providerURL:        "registry.terraform.io/hashicorp/aws/5.0.1",
			db:                 db,
			expectedID:         "hashicorp:aws:aws_db_instance:testmysql",
			expectedRes: apiv1.Resource{
				ID:   "hashicorp:aws:aws_db_instance:testmysql",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"allocated_storage":   20,
					"engine":              "mysql",
					"engine_version":      "8.0",
					"identifier":          "testmysql",
					"instance_class":      "db.t3.micro",
					"password":            "$kusion_path.hashicorp:random:random_password:testmysql-mysql.result",
					"publicly_accessible": true,
					"skip_final_snapshot": true,
					"username":            "root",
					"vpc_security_group_ids": []string{
						"$kusion_path.hashicorp:aws:aws_security_group:testmysql-mysql.id",
					},
				},
				Extensions: map[string]interface{}{
					"provider": "registry.terraform.io/hashicorp/aws/5.0.1",
					"providerMeta": map[string]interface{}{
						"region": "us-east-1",
					},
					"resourceType": awsDBInstance,
				},
			},
		},
	}

	for _, test := range tests {
		var awsProvider = &inputs.Provider{}
		_ = awsProvider.SetString(test.providerURL)
		actualID, actualRes := g.(*mysqlGenerator).generateAWSDBInstance(
			test.region, test.awsSecurityGroupID, test.randomPasswordID,
			awsProvider, test.db,
		)
		assert.Equal(t, test.expectedID, actualID)
		assert.Equal(t, test.expectedRes, actualRes)
	}
}
