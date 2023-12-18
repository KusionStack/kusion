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
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/container"
)

func TestNewMySQLGenerator(t *testing.T) {
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
	mysql := &mysql.MySQL{}
	ws := &workspaceapi.Workspace{}
	generator, err := NewMySQLGenerator(project, stack, appName, workload, mysql, ws)

	assert.NoError(t, err)
	assert.NotNil(t, generator)
}

func TestGenerate(t *testing.T) {
	g := genCloudMySQLGenerator()

	spec := &intent.Intent{}
	err := g.Generate(spec)

	var providerMeta map[string]interface{}
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
					"provider":     "registry.terraform.io/hashicorp/random/3.5.1",
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
					"resourceType": "aws_security_group",
				},
			},
			intent.Resource{
				ID:   "hashicorp:aws:aws_db_instance:testapp",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"allocated_storage":   g.mysql.Size,
					"engine":              dbEngine,
					"engine_version":      g.mysql.Version,
					"identifier":          g.appName,
					"instance_class":      g.mysql.InstanceType,
					"password":            "$kusion_path.hashicorp:random:random_password:testapp-db.result",
					"publicly_accessible": true,
					"skip_final_snapshot": true,
					"username":            g.mysql.Username,
					"vpc_security_group_ids": []string{
						"$kusion_path.hashicorp:aws:aws_security_group:testapp-db.id",
					},
				},
				Extensions: map[string]interface{}{
					"provider": "registry.terraform.io/hashicorp/aws/5.0.1",
					"providerMeta": map[string]interface{}{
						"region": "us-east-1",
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
						"username":    g.mysql.Username,
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

func TestPatchWorkspaceConfig(t *testing.T) {
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
		Type:    "cloud",
		Version: "5.7",
	}
	ws := &workspaceapi.Workspace{
		Name: "testworkspace",
		Runtimes: &workspaceapi.RuntimeConfigs{
			Kubernetes: &workspaceapi.KubernetesConfig{
				KubeConfig: "/Users/username/testkubeconfig",
			},
			Terraform: workspaceapi.TerraformConfig{
				"random": &workspaceapi.ProviderConfig{
					Source:  "hashicorp/random",
					Version: "3.5.1",
				},
				"aws": &workspaceapi.ProviderConfig{
					Source:  "hashicorp/aws",
					Version: "5.0.1",
					GenericConfig: workspaceapi.GenericConfig{
						"region": "us-east-1",
					},
				},
			},
		},
		Modules: workspaceapi.ModuleConfigs{
			"mysql": &workspaceapi.ModuleConfig{
				Default: workspaceapi.GenericConfig{
					"cloud":          "aws",
					"size":           20,
					"instanceType":   "db.t3.micro",
					"privateRouting": false,
				},
			},
		},
	}

	g := &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}
	err := g.patchWorkspaceConfig()

	assert.NoError(t, err)
	assert.Equal(t, g.mysql, genCloudMySQLGenerator().mysql)
}

func TestGetTFProviderType(t *testing.T) {
	g := genCloudMySQLGenerator()
	providerType, err := g.getTFProviderType()

	assert.NoError(t, err)
	assert.Equal(t, providerType, "aws")
}

func TestInjectSecret(t *testing.T) {
	g := genCloudMySQLGenerator()
	g.workload = &workload.Workload{
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

	data := make(map[string]string)
	data["hostAddress"] = "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
	data["username"] = g.mysql.Username
	data["password"] = "$kusion_path.hashicorp:random:random_password:testapp-db.result"

	// Create the k8s secret and append to the spec.
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.appName + dbResSuffix,
			Namespace: g.project.Name,
		},
		StringData: data,
	}

	err := g.injectSecret(secret)
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
	assert.Equal(t, expectedContainer, g.workload.Service.Containers["testcontainer"])
}

func TestGenerateTFRandomPassword(t *testing.T) {
	g := genCloudMySQLGenerator()

	randomProvider := &inputs.Provider{}
	randomProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.RandomProvider])
	_ = randomProvider.SetString(randomProviderURL)

	var providerMeta map[string]interface{}
	id, res := g.generateTFRandomPassword(randomProvider)
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
	g := genCloudMySQLGenerator()

	spec := &intent.Intent{}
	hostAddress := "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
	username := g.mysql.Username
	password := "$kusion_path.hashicorp:random:random_password:testapp-db.result"
	secret, err := g.generateDBSecret(hostAddress, username, password, spec)

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
			Name:      g.appName + dbResSuffix,
			Namespace: g.project.Name,
		},
		StringData: data,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func genCloudMySQLGenerator() *mysqlGenerator {
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
		Type:           "cloud",
		Version:        "5.7",
		Size:           20,
		InstanceType:   "db.t3.micro",
		PrivateRouting: false,
		Username:       defaultUsername,
		Category:       defaultCategory,
		SecurityIPs:    defaultSecurityIPs,
	}
	ws := &workspaceapi.Workspace{
		Name: "testworkspace",
		Runtimes: &workspaceapi.RuntimeConfigs{
			Kubernetes: &workspaceapi.KubernetesConfig{
				KubeConfig: "/Users/username/testkubeconfig",
			},
			Terraform: workspaceapi.TerraformConfig{
				"random": &workspaceapi.ProviderConfig{
					Source:  "hashicorp/random",
					Version: "3.5.1",
				},
				"aws": &workspaceapi.ProviderConfig{
					Source:  "hashicorp/aws",
					Version: "5.0.1",
					GenericConfig: workspaceapi.GenericConfig{
						"region": "us-east-1",
					},
				},
			},
		},
		Modules: workspaceapi.ModuleConfigs{
			"mysql": &workspaceapi.ModuleConfig{
				Default: workspaceapi.GenericConfig{
					"cloud":          "aws",
					"size":           20,
					"instanceType":   "db.t3.micro",
					"privateRouting": false,
				},
			},
		},
	}

	return &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}
}
