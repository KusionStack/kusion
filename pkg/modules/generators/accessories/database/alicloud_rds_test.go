package accessories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/database"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestGenerateAlicloudResources(t *testing.T) {
	project := &apiv1.Project{
		Name: "testproject",
	}
	stack := &apiv1.Stack{
		Name: "teststack",
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:         "alicloud",
		Engine:       "MySQL",
		Version:      "5.7",
		Size:         20,
		InstanceType: "mysql.n2.serverless.1c",
		Category:     "serverless_basic",
		Username:     "root",
		SecurityIPs: []string{
			"0.0.0.0/0",
		},
		PrivateRouting: true,
		SubnetID:       "test_subnet_id",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	alicloudProviderRegion = "cn-beijing"
	spec := &intent.Intent{}
	secret, err := generator.generateAlicloudResources(database, spec)

	hostAddress := "$kusion_path.aliyun:alicloud:alicloud_db_instance:testapp.connection_string"
	username := database.Username
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

func TestGenerateAlicloudDBInstance(t *testing.T) {
	alicloudProviderRegion = "cn-beijing"
	alicloudProvider := &inputs.Provider{}
	alicloudProvider.SetString(defaultAlicloudProvider)

	project := &apiv1.Project{
		Name: "testproject",
	}
	stack := &apiv1.Stack{
		Name: "teststack",
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:         "alicloud",
		Engine:       "MySQL",
		Version:      "5.7",
		Size:         20,
		InstanceType: "mysql.n2.serverless.1c",
		Category:     "serverless_basic",
		Username:     "root",
		SecurityIPs: []string{
			"0.0.0.0/0",
		},
		PrivateRouting: true,
		SubnetID:       "test_subnet_id",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	alicloudDBInstanceID, r := generator.generateAlicloudDBInstance(alicloudProviderRegion, alicloudProvider, database)
	expectedAlicloudDBInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	expectedRes := intent.Resource{
		ID:   "aliyun:alicloud:alicloud_db_instance:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"category":                 database.Category,
			"db_instance_storage_type": "cloud_essd",
			"engine":                   database.Engine,
			"engine_version":           database.Version,
			"instance_charge_type":     "Serverless",
			"instance_storage":         database.Size,
			"instance_type":            database.InstanceType,
			"security_ips":             database.SecurityIPs,
			"serverless_config": []alicloudServerlessConfig{
				{
					AutoPause:   false,
					SwitchForce: false,
					MaxCapacity: 8,
					MinCapacity: 1,
				},
			},
			"vswitch_id": database.SubnetID,
		},
		Extensions: map[string]interface{}{
			"provider": defaultAlicloudProvider,
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
	alicloudProviderRegion = "cn-beijing"
	alicloudProvider := &inputs.Provider{}
	alicloudProvider.SetString(defaultAlicloudProvider)

	project := &apiv1.Project{
		Name: "testproject",
	}
	stack := &apiv1.Stack{
		Name: "teststack",
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:         "alicloud",
		Engine:       "MySQL",
		Version:      "5.7",
		Size:         20,
		InstanceType: "mysql.n2.serverless.1c",
		Category:     "serverless_basic",
		Username:     "root",
		SecurityIPs: []string{
			"0.0.0.0/0",
		},
		PrivateRouting: true,
		SubnetID:       "test_subnet_id",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	dbInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	alicloudDBConnectionID, r := generator.generateAlicloudDBConnection(dbInstanceID, alicloudProviderRegion, alicloudProvider)
	expectedAlicloudDBConnectionID := "aliyun:alicloud:alicloud_db_connection:testapp"
	expectedRes := intent.Resource{
		ID:   "aliyun:alicloud:alicloud_db_connection:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"instance_id": "$kusion_path.aliyun:alicloud:alicloud_db_instance:testapp.id",
		},
		Extensions: map[string]interface{}{
			"provider": defaultAlicloudProvider,
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
	alicloudProviderRegion = "cn-beijing"
	alicloudProvider := &inputs.Provider{}
	alicloudProvider.SetString(defaultAlicloudProvider)

	project := &apiv1.Project{
		Name: "testproject",
	}
	stack := &apiv1.Stack{
		Name: "teststack",
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:         "alicloud",
		Engine:       "MySQL",
		Version:      "5.7",
		Size:         20,
		InstanceType: "mysql.n2.serverless.1c",
		Category:     "serverless_basic",
		Username:     "root",
		SecurityIPs: []string{
			"0.0.0.0/0",
		},
		PrivateRouting: true,
		SubnetID:       "test_subnet_id",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	accountName := database.Username
	randomPasswordID := "hashicorp:random:random_password:testapp-db"
	alicloudDBInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	r := generator.generateAlicloudRDSAccount(accountName, randomPasswordID, alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider, database)

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
			"provider": defaultAlicloudProvider,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_rds_account",
		},
	}

	assert.Equal(t, expectedRes, r)
}
