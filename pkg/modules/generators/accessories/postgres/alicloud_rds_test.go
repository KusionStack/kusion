package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules/inputs"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestPostgresGenerator_GenerateAlicloudResources(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": {
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"postgres": {
			"cloud":          "alicloud",
			"size":           20,
			"instanceType":   "pg.n2.serverless.1c",
			"category":       "serverless_basic",
			"privateRouting": false,
			"subnetID":       "xxxxxxx",
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"alicloud": &apiv1.ProviderConfig{
			Version: "1.209.1",
			Source:  "aliyun/alicloud",
			GenericConfig: apiv1.GenericConfig{
				"region": "cn-beijing",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "pg.n2.serverless.1c",
		Category:       "serverless_basic",
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		SubnetID:       "xxxxxxx",
		DatabaseName:   "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name        string
		db          *postgres.PostgreSQL
		spec        *apiv1.Intent
		expected    *v1.Secret
		expectedErr error
	}{
		{
			name: "Generate Alicloud Resources",
			db:   db,
			spec: &apiv1.Intent{},
			expected: &v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: v1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testpostgres-postgres",
					Namespace: "testproject",
				},
				StringData: map[string]string{
					"hostAddress": "$kusion_path.aliyun:alicloud:alicloud_db_connection:testpostgres.connection_string",
					"username":    "root",
					"password":    "$kusion_path.hashicorp:random:random_password:testpostgres-postgres.result",
				},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*postgresGenerator).generateAlicloudResources(test.db, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestPostgresGenerator_GenerateAlicloudDBInstance(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": {
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"postgres": {
			"cloud":          "alicloud",
			"size":           20,
			"instanceType":   "pg.n2.serverless.1c",
			"category":       "serverless_basic",
			"privateRouting": false,
			"subnetID":       "xxxxxxx",
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"alicloud": &apiv1.ProviderConfig{
			Version: "1.209.1",
			Source:  "aliyun/alicloud",
			GenericConfig: apiv1.GenericConfig{
				"region": "cn-beijing",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "pg.n2.serverless.1c",
		Category:       "serverless_basic",
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		SubnetID:       "xxxxxxx",
		DatabaseName:   "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name        string
		region      string
		providerURL string
		db          *postgres.PostgreSQL
		expectedID  string
		expectedRes apiv1.Resource
	}{
		{
			name:        "Generate Alicloud DB Instance",
			region:      "cn-beijing",
			providerURL: "registry.terraform.io/aliyun/alicloud/1.209.1",
			db:          db,
			expectedID:  "aliyun:alicloud:alicloud_db_instance:testpostgres",
			expectedRes: apiv1.Resource{
				ID:   "aliyun:alicloud:alicloud_db_instance:testpostgres",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"category":                 "serverless_basic",
					"db_instance_storage_type": "cloud_essd",
					"engine":                   "PostgreSQL",
					"engine_version":           "8.0",
					"instance_name":            "testpostgres",
					"instance_charge_type":     "Serverless",
					"instance_storage":         20,
					"instance_type":            "pg.n2.serverless.1c",
					"security_ips": []string{
						"0.0.0.0/0",
					},
					"serverless_config": []alicloudServerlessConfig{
						{
							AutoPause:   false,
							SwitchForce: false,
							MaxCapacity: 12,
							MinCapacity: 1,
						},
					},
					"vswitch_id": "xxxxxxx",
				},
				Extensions: map[string]interface{}{
					"provider": "registry.terraform.io/aliyun/alicloud/1.209.1",
					"providerMeta": map[string]interface{}{
						"region": "cn-beijing",
					},
					"resourceType": alicloudDBInstance,
				},
			},
		},
	}

	for _, test := range tests {
		alicloudProvider := &inputs.Provider{}
		_ = alicloudProvider.SetString(test.providerURL)
		actualID, actualRes := g.(*postgresGenerator).generateAlicloudDBInstance(
			test.region, alicloudProvider, test.db,
		)

		assert.Equal(t, test.expectedID, actualID)
		assert.Equal(t, test.expectedRes, actualRes)
	}
}

func TestPostgresGenerator_GenerateAlicloudDBConnection(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": {
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"postgres": {
			"cloud":          "alicloud",
			"size":           20,
			"instanceType":   "pg.n2.serverless.1c",
			"category":       "serverless_basic",
			"privateRouting": false,
			"subnetID":       "xxxxxxx",
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"alicloud": &apiv1.ProviderConfig{
			Version: "1.209.1",
			Source:  "aliyun/alicloud",
			GenericConfig: apiv1.GenericConfig{
				"region": "cn-beijing",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "pg.n2.serverless.1c",
		Category:       "serverless_basic",
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		SubnetID:       "xxxxxxx",
		DatabaseName:   "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name         string
		dbInstanceID string
		region       string
		providerURL  string
		expectedID   string
		expectedRes  apiv1.Resource
	}{
		{
			name:         "Generate Alicloud DB Connection",
			dbInstanceID: "aliyun:alicloud:alicloud_db_instance:testpostgres",
			region:       "cn-beijing",
			providerURL:  "registry.terraform.io/aliyun/alicloud/1.209.1",
			expectedID:   "aliyun:alicloud:alicloud_db_connection:testpostgres",
			expectedRes: apiv1.Resource{
				ID:   "aliyun:alicloud:alicloud_db_connection:testpostgres",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"instance_id": "$kusion_path.aliyun:alicloud:alicloud_db_instance:testpostgres.id",
				},
				Extensions: map[string]interface{}{
					"provider": "registry.terraform.io/aliyun/alicloud/1.209.1",
					"providerMeta": map[string]interface{}{
						"region": "cn-beijing",
					},
					"resourceType": alicloudDBConnection,
				},
			},
		},
	}

	for _, test := range tests {
		alicloudProvider := &inputs.Provider{}
		_ = alicloudProvider.SetString(test.providerURL)
		actualID, actualRes := g.(*postgresGenerator).generateAlicloudDBConnection(
			test.dbInstanceID, test.region, alicloudProvider,
		)

		assert.Equal(t, test.expectedID, actualID)
		assert.Equal(t, test.expectedRes, actualRes)
	}
}

func TestPostgresGenerator_GenerateAlicloudDBAccount(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": {
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "cloud",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		"postgres": {
			"cloud":          "alicloud",
			"size":           20,
			"instanceType":   "pg.n2.serverless.1c",
			"category":       "serverless_basic",
			"privateRouting": false,
			"subnetID":       "xxxxxxx",
		},
	}
	tfConfigs := apiv1.TerraformConfig{
		"random": &apiv1.ProviderConfig{
			Version: "3.5.1",
			Source:  "hashicorp/random",
		},
		"alicloud": &apiv1.ProviderConfig{
			Version: "1.209.1",
			Source:  "aliyun/alicloud",
			GenericConfig: apiv1.GenericConfig{
				"region": "cn-beijing",
			},
		},
	}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:           "cloud",
		Version:        "8.0",
		Size:           20,
		InstanceType:   "pg.n2.serverless.1c",
		Category:       "serverless_basic",
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
		PrivateRouting: false,
		SubnetID:       "xxxxxxx",
		DatabaseName:   "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name             string
		providerURL      string
		accountName      string
		randomPasswordID string
		dbInstanceID     string
		region           string
		db               *postgres.PostgreSQL
		expectedRes      apiv1.Resource
	}{
		{
			name:             "Generate Alicloud RDS Account",
			providerURL:      "registry.terraform.io/aliyun/alicloud/1.209.1",
			accountName:      "root",
			randomPasswordID: "hashicorp:random:random_password:testpostgres-postgres",
			dbInstanceID:     "aliyun:alicloud:alicloud_db_instance:testpostgres",
			region:           "cn-beijing",
			db:               db,
			expectedRes: apiv1.Resource{
				ID:   "aliyun:alicloud:alicloud_rds_account:testpostgres",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"account_name":     "root",
					"account_password": "$kusion_path.hashicorp:random:random_password:testpostgres-postgres.result",
					"account_type":     "Super",
					"db_instance_id":   "$kusion_path.aliyun:alicloud:alicloud_db_instance:testpostgres.id",
				},
				Extensions: map[string]interface{}{
					"provider": "registry.terraform.io/aliyun/alicloud/1.209.1",
					"providerMeta": map[string]interface{}{
						"region": "cn-beijing",
					},
					"resourceType": alicloudRDSAccount,
				},
			},
		},
	}

	for _, test := range tests {
		alicloudProvider := &inputs.Provider{}
		_ = alicloudProvider.SetString(test.providerURL)
		actualRes := g.(*postgresGenerator).generateAlicloudRDSAccount(
			test.accountName, test.randomPasswordID, test.dbInstanceID, test.region, alicloudProvider, test.db)

		assert.Equal(t, test.expectedRes, actualRes)
	}
}
