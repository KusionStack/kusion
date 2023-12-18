package accessories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

var (
	defaultAWSProvider = ""
	awsProviderRegion  = ""
)

func TestGenerateAWSResources(t *testing.T) {
	project := &project.Project{
		Configuration: project.Configuration{
			Name: "testproject",
		},
	}
	stack := &stack.Stack{
		Configuration: stack.Configuration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	mysql := &mysql.MySQL{
		Type:         "aws",
		Version:      "5.7",
		InstanceType: "db.t3.micro",
		Size:         10,
		Username:     "root",
	}
	ws := &workspaceapi.Workspace{}

	generator := &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}

	spec := &intent.Intent{}
	secret, err := generator.generateAWSResources(mysql, spec)

	hostAddress := "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
	username := mysql.Username
	password := "$kusion_path.hashicorp:random:random_password:testapp-db.result"
	data := make(map[string]string)
	data["hostAddress"] = hostAddress
	data["username"] = username
	data["password"] = password
	expectedSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName + dbResSuffix,
			Namespace: project.Name,
		},
		StringData: data,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func TestGenerateAWSSecurityGroup(t *testing.T) {
	awsProvider := &inputs.Provider{}

	project := &project.Project{
		Configuration: project.Configuration{
			Name: "testproject",
		},
	}
	stack := &stack.Stack{
		Configuration: stack.Configuration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	mysql := &mysql.MySQL{
		Type:         "aws",
		Version:      "5.7",
		InstanceType: "db.t3.micro",
		Size:         10,
		Username:     "root",
	}
	ws := &workspaceapi.Workspace{}

	generator := &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}

	var cidrBlocks []string
	awsSecurityGroupID, r, err := generator.generateAWSSecurityGroup(awsProvider, awsProviderRegion, mysql)
	expectedAWSSecurityGroupID := "hashicorp:aws:aws_security_group:testapp-db"
	expectedRes := intent.Resource{
		ID:   "hashicorp:aws:aws_security_group:testapp-db",
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
					CidrBlocks: cidrBlocks,
					Protocol:   "tcp",
					FromPort:   3306,
					ToPort:     3306,
				},
			},
		},
		Extensions: map[string]interface{}{
			"provider": defaultAWSProvider,
			"providerMeta": map[string]interface{}{
				"region": awsProviderRegion,
			},
			"resourceType": "aws_security_group",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedAWSSecurityGroupID, awsSecurityGroupID)
	assert.Equal(t, expectedRes, r)
}

func TestGenerateAWSDBInstance(t *testing.T) {
	awsProvider := &inputs.Provider{}
	awsProvider.SetString(defaultAWSProvider)
	awsProviderRegion = "us-east-1"

	project := &project.Project{
		Configuration: project.Configuration{
			Name: "testproject",
		},
	}
	stack := &stack.Stack{
		Configuration: stack.Configuration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	mysql := &mysql.MySQL{
		Type:         "aws",
		Version:      "5.7",
		InstanceType: "db.t3.micro",
		Size:         10,
		Username:     "root",
	}
	ws := &workspaceapi.Workspace{}

	generator := &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}

	awsSecurityGroupID := "hashicorp:aws:aws_security_group:testapp-db"
	randomPasswordID := "hashicorp:random:random_password:testapp-db"

	awsDBInstanceID, r := generator.generateAWSDBInstance(awsProviderRegion, awsSecurityGroupID, randomPasswordID, awsProvider, mysql)
	expectedAWSDBInstanceID := "hashicorp:aws:aws_db_instance:testapp"
	expectedRes := intent.Resource{
		ID:   "hashicorp:aws:aws_db_instance:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"allocated_storage":   mysql.Size,
			"engine":              dbEngine,
			"engine_version":      mysql.Version,
			"identifier":          appName,
			"instance_class":      mysql.InstanceType,
			"password":            "$kusion_path.hashicorp:random:random_password:testapp-db.result",
			"publicly_accessible": false,
			"skip_final_snapshot": true,
			"username":            mysql.Username,
			"vpc_security_group_ids": []string{
				"$kusion_path.hashicorp:aws:aws_security_group:testapp-db.id",
			},
		},
		Extensions: map[string]interface{}{
			"provider": defaultAWSProvider,
			"providerMeta": map[string]interface{}{
				"region": awsProviderRegion,
			},
			"resourceType": "aws_db_instance",
		},
	}

	assert.Equal(t, expectedAWSDBInstanceID, awsDBInstanceID)
	assert.Equal(t, expectedRes, r)
}
