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

func TestGenerateAlicloudResources(t *testing.T) {
	g := genAlicloudMySQLGenerator()

	spec := &intent.Intent{}
	secret, err := g.generateAlicloudResources(g.mysql, spec)

	hostAddress := "$kusion_path.aliyun:alicloud:alicloud_db_connection:testapp.connection_string"
	username := g.mysql.Username
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
			Name:      g.appName + dbResSuffix,
			Namespace: g.project.Name,
		},
		StringData: data,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func TestGenerateAlicloudDBInstance(t *testing.T) {
	g := genAlicloudMySQLGenerator()
	alicloudProvider := &inputs.Provider{}
	alicloudProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])
	_ = alicloudProvider.SetString(alicloudProviderURL)
	alicloudProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])

	alicloudDBInstanceID, r := g.generateAlicloudDBInstance(alicloudProviderRegion, alicloudProvider, g.mysql)
	expectedAlicloudDBInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	expectedRes := intent.Resource{
		ID:   "aliyun:alicloud:alicloud_db_instance:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"category":                 g.mysql.Category,
			"db_instance_storage_type": "cloud_essd",
			"engine":                   dbEngine,
			"engine_version":           g.mysql.Version,
			"instance_charge_type":     "Serverless",
			"instance_storage":         g.mysql.Size,
			"instance_type":            g.mysql.InstanceType,
			"security_ips":             g.mysql.SecurityIPs,
			"serverless_config": []alicloudServerlessConfig{
				{
					AutoPause:   false,
					SwitchForce: false,
					MaxCapacity: 8,
					MinCapacity: 1,
				},
			},
			"vswitch_id": g.mysql.SubnetID,
		},
		Extensions: map[string]interface{}{
			"provider": alicloudProviderURL,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_db_instance",
		},
	}

	assert.Equal(t, expectedAlicloudDBInstanceID, alicloudDBInstanceID)
	assert.Equal(t, expectedRes, r)
}

func TestGenerateAlicloudDBConnection(t *testing.T) {
	g := genAlicloudMySQLGenerator()
	alicloudProvider := &inputs.Provider{}
	alicloudProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])
	_ = alicloudProvider.SetString(alicloudProviderURL)
	alicloudProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])

	dbInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	alicloudDBConnectionID, r := g.generateAlicloudDBConnection(dbInstanceID, alicloudProviderRegion, alicloudProvider)
	expectedAlicloudDBConnectionID := "aliyun:alicloud:alicloud_db_connection:testapp"
	expectedRes := intent.Resource{
		ID:   "aliyun:alicloud:alicloud_db_connection:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"instance_id": "$kusion_path.aliyun:alicloud:alicloud_db_instance:testapp.id",
		},
		Extensions: map[string]interface{}{
			"provider": alicloudProviderURL,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_db_connection",
		},
	}

	assert.Equal(t, expectedAlicloudDBConnectionID, alicloudDBConnectionID)
	assert.Equal(t, expectedRes, r)
}

func TestGenerateAlicloudRDSAccount(t *testing.T) {
	g := genAlicloudMySQLGenerator()
	alicloudProvider := &inputs.Provider{}
	alicloudProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])
	_ = alicloudProvider.SetString(alicloudProviderURL)
	alicloudProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])

	accountName := g.mysql.Username
	randomPasswordID := "hashicorp:random:random_password:testapp-db"
	alicloudDBInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	r := g.generateAlicloudRDSAccount(accountName, randomPasswordID, alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider, g.mysql)

	expectedRes := intent.Resource{
		ID:   "aliyun:alicloud:alicloud_rds_account:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"account_name":     accountName,
			"account_password": "$kusion_path.hashicorp:random:random_password:testapp-db.result",
			"account_type":     "Super",
			"db_instance_id":   "$kusion_path.aliyun:alicloud:alicloud_db_instance:testapp.id",
		},
		Extensions: map[string]interface{}{
			"provider": alicloudProviderURL,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_rds_account",
		},
	}

	assert.Equal(t, expectedRes, r)
}

func genAlicloudMySQLGenerator() *mysqlGenerator {
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
		InstanceType:   "mysql.n2.serverless.1c",
		Category:       "serverless_basic",
		PrivateRouting: false,
		SubnetID:       "test_subnet_id",
		Username:       defaultUsername,
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
				"alicloud": &workspaceapi.ProviderConfig{
					Source:  "aliyun/alicloud",
					Version: "1.209.1",
					GenericConfig: workspaceapi.GenericConfig{
						"region": "cn-beijing",
					},
				},
			},
		},
		Modules: workspaceapi.ModuleConfigs{
			"mysql": &workspaceapi.ModuleConfig{
				Default: workspaceapi.GenericConfig{
					"cloud":          "alicloud",
					"size":           20,
					"instanceType":   "mysql.n2.serverless.1c",
					"category":       "serverless_basic",
					"privateRouting": false,
					"subnetID":       "test_subnet_id",
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
