package accessories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/database"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/container"
)

func TestNewDatabaseGenerator(t *testing.T) {
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
	database := &database.Database{}
	generator, err := NewDatabaseGenerator(project, stack, appName, workload, database)

	assert.NoError(t, err)
	assert.NotNil(t, generator)
}

func TestGenerate(t *testing.T) {
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
	database := &database.Database{
		Type:         "aws",
		Engine:       "mysql",
		Version:      "5.7",
		InstanceType: "db.t3.micro",
		Size:         10,
		Username:     "root",
	}
	generator, _ := NewDatabaseGenerator(project, stack, appName, workload, database)

	awsProviderRegion = "us-east-1"
	spec := &intent.Intent{}
	err := generator.Generate(spec)

	var providerMeta map[string]interface{}
	var cidrBlocks []string
	expectedSpec := &intent.Intent{
		Resources: intent.Resources{
			intent.Resource{
				ID:   "hashicorp:random:random_password:testapp-db",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"length":           16,
					"override_special": "_",
					"special":          true,
				},
				Extensions: map[string]interface{}{
					"provider":     randomProviderURL,
					"providerMeta": providerMeta,
					"resourceType": "random_password",
				},
			},
			intent.Resource{
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
			},
			intent.Resource{
				ID:   "hashicorp:aws:aws_db_instance:testapp",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"allocated_storage":   database.Size,
					"engine":              database.Engine,
					"engine_version":      database.Version,
					"identifier":          appName,
					"instance_class":      database.InstanceType,
					"password":            "$kusion_path.hashicorp:random:random_password:testapp-db.result",
					"publicly_accessible": false,
					"skip_final_snapshot": true,
					"username":            database.Username,
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
			},
			intent.Resource{
				ID:   "v1:Secret:testproject:testapp-db",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"name":              "testapp-db",
						"namespace":         "testproject",
					},
					"stringData": map[string]interface{}{
						"hostAddress": "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address",
						"password":    "$kusion_path.hashicorp:random:random_password:testapp-db.result",
						"username":    database.Username,
					},
				},
				Extensions: map[string]interface{}{
					"GVK": "/v1, Kind=Secret",
				},
			},
		},
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Resources)
	assert.Equal(t, expectedSpec.Resources, spec.Resources)
}

func TestInjectSecret(t *testing.T) {
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
	workload := &workload.Workload{
		Service: &workload.Service{
			Base: workload.Base{
				Containers: map[string]container.Container{
					"testcontainer": {
						Image: "testimage:latest",
					},
				},
			},
		},
	}
	database := &database.Database{
		Type:         "aws",
		Engine:       "mysql",
		Version:      "5.7",
		InstanceType: "db.t3.micro",
		Size:         10,
		Username:     "root",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	data := make(map[string]string)
	data["hostAddress"] = "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
	data["username"] = database.Username
	data["password"] = "$kusion_path.hashicorp:random:random_password:testapp-db.result"

	// Create the k8s secret and append to the spec.
	secret := &v1.Secret{
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

	err := generator.injectSecret(secret)
	expectedContainer := container.Container{
		Image: "testimage:latest",
		Env: yaml.MapSlice{
			{
				Key:   dbHostAddressEnv,
				Value: "secret://" + secret.Name + "/hostAddress",
			},
			{
				Key:   dbUsernameEnv,
				Value: "secret://" + secret.Name + "/username",
			},
			{
				Key:   dbPasswordEnv,
				Value: "secret://" + secret.Name + "/password",
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedContainer, workload.Service.Containers["testcontainer"])
}

func TestGenerateTFRandomPassword(t *testing.T) {
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
	database := &database.Database{}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}
	randomProvider := &inputs.Provider{}
	randomProvider.SetString(randomProviderURL)

	var providerMeta map[string]interface{}
	id, res := generator.generateTFRandomPassword(randomProvider)
	expectedID := "hashicorp:random:random_password:testapp-db"
	expectedRes := intent.Resource{
		ID:   "hashicorp:random:random_password:testapp-db",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"length":           16,
			"override_special": "_",
			"special":          true,
		},
		Extensions: map[string]interface{}{
			"provider":     randomProviderURL,
			"providerMeta": providerMeta,
			"resourceType": "random_password",
		},
	}

	assert.Equal(t, expectedID, id)
	assert.Equal(t, expectedRes, res)
}

func TestGenerateDBSeret(t *testing.T) {
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
	database := &database.Database{}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	spec := &intent.Intent{}
	hostAddress := "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
	username := database.Username
	password := "$kusion_path.hashicorp:random:random_password:testapp-db.result"
	secret, err := generator.generateDBSecret(hostAddress, username, password, spec)

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
