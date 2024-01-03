package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/container"
	"kusionstack.io/kusion/pkg/workspace"
)

func newGeneratorContext(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	workload *workload.Workload,
	database map[string]*database.Database,
	moduleInputs map[string]apiv1.GenericConfig,
	tfConfigs apiv1.TerraformConfig,
) modules.GeneratorContext {
	application := &inputs.AppConfiguration{
		Name:     appName,
		Workload: workload,
		Database: database,
	}

	return modules.GeneratorContext{
		Project:         project,
		Stack:           stack,
		Application:     application,
		Namespace:       project.Name,
		ModuleInputs:    moduleInputs,
		TerraformConfig: tfConfigs,
	}
}

func TestNewMySQLGenerator(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": {
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
		"mysql": {
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
		Type:    "cloud",
		Version: "8.0",
	}

	tests := []struct {
		name        string
		ctx         modules.GeneratorContext
		dbKey       string
		db          *mysql.MySQL
		expected    modules.Generator
		expectedErr error
	}{
		{
			name:  "New Valid MySQL Generator",
			ctx:   context,
			dbKey: "testmysql",
			db:    db,
			expected: &mysqlGenerator{
				project:       project,
				stack:         stack,
				appName:       appName,
				workload:      workload,
				mysql:         db,
				moduleConfigs: moduleInputs,
				tfConfigs:     tfConfigs,
				namespace:     project.Name,
				dbKey:         "testmysql",
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := NewMySQLGenerator(test.ctx, test.dbKey, test.db)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestNewMySQLGeneratorFunc(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": {
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
		"mysql": {
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
		Type:    "cloud",
		Version: "8.0",
	}

	tests := []struct {
		name        string
		ctx         modules.GeneratorContext
		dbKey       string
		db          *mysql.MySQL
		expected    modules.Generator
		expectedErr error
	}{
		{
			name:  "New Valid MySQL Generator Func",
			ctx:   context,
			dbKey: "testmysql",
			db:    db,
			expected: &mysqlGenerator{
				project:       project,
				stack:         stack,
				appName:       appName,
				workload:      workload,
				mysql:         db,
				moduleConfigs: moduleInputs,
				tfConfigs:     tfConfigs,
				namespace:     project.Name,
				dbKey:         "testmysql",
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		testGeneratorFunc := NewMySQLGeneratorFunc(test.ctx, test.dbKey, test.db)
		actual, actualErr := testGeneratorFunc()
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_Generate(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
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
		"alicloud": &apiv1.ProviderConfig{
			Version: "1.209.1",
			Source:  "aliyun/alicloud",
			GenericConfig: apiv1.GenericConfig{
				"region": "cn-beijing",
			},
		},
	}

	tests := []struct {
		name         string
		database     map[string]*database.Database
		moduleInputs map[string]apiv1.GenericConfig
		db           *mysql.MySQL
		expectedErr  error
	}{
		{
			name: "Generate Local Database",
			database: map[string]*database.Database{
				"testmysql": {
					Header: database.Header{
						Type: "MySQL",
					},
					MySQL: &mysql.MySQL{
						Type:    "local",
						Version: "8.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{},
			db: &mysql.MySQL{
				Type:    "local",
				Version: "8.0",
			},
			expectedErr: nil,
		},
		{
			name: "Generate AWS RDS",
			database: map[string]*database.Database{
				"testmysql": {
					Header: database.Header{
						Type: "MySQL",
					},
					MySQL: &mysql.MySQL{
						Type:    "cloud",
						Version: "8.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud":          "aws",
					"size":           20,
					"instanceType":   "db.t3.micro",
					"privateRouting": false,
				},
			},
			db: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
			expectedErr: nil,
		},
		{
			name: "Generate Alicloud RDS",
			database: map[string]*database.Database{
				"testmysql": {
					Header: database.Header{
						Type: "MySQL",
					},
					MySQL: &mysql.MySQL{
						Type:    "cloud",
						Version: "8.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud":          "alicloud",
					"size":           20,
					"instanceType":   "mysql.n2.serverless.1c",
					"category":       "serverless_basic",
					"privateRouting": false,
					"subnetID":       "xxxxxxx",
				},
			},
			db: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
			expectedErr: nil,
		},
		{
			name: "Empty Cloud MySQL Instance Type",
			database: map[string]*database.Database{
				"testmysql": {
					Header: database.Header{
						Type: "MySQL",
					},
					MySQL: &mysql.MySQL{
						Type:    "cloud",
						Version: "8.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud": "alicloud",
				},
			},
			db: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
			expectedErr: fmt.Errorf(mysql.ErrEmptyInstanceTypeForCloudDB),
		},
		{
			name: "Empty Cloud MySQL Instance Type",
			database: map[string]*database.Database{
				"testmysql": {
					Header: database.Header{
						Type: "MySQL",
					},
					MySQL: &mysql.MySQL{
						Type:    "cloud",
						Version: "8.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud":        "unsupported-type",
					"instanceType": "db.t3.micro",
				},
			},
			db: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
			expectedErr: fmt.Errorf(errUnsupportedTFProvider, "unsupported-type"),
		},
	}

	for _, test := range tests {
		context := newGeneratorContext(project, stack, appName, workload, test.database,
			test.moduleInputs, tfConfigs)
		g, _ := NewMySQLGenerator(context, "testmysql", test.db)
		actualErr := g.(*mysqlGenerator).Generate(&apiv1.Intent{})

		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_PatchWorkspaceConfig(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": {
			Header: database.Header{
				Type: "MySQL",
			},
			MySQL: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
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
	db := &mysql.MySQL{
		Type:    "cloud",
		Version: "8.0",
	}

	tests := []struct {
		name         string
		moduleInputs map[string]apiv1.GenericConfig
		expected     *mysql.MySQL
		expectedErr  error
	}{
		{
			name: "MySQL with Default Values",
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud":        "aws",
					"instanceType": "db.t3.micro",
				},
			},
			expected: &mysql.MySQL{
				Type:           "cloud",
				Version:        "8.0",
				Size:           defaultSize,
				InstanceType:   "db.t3.micro",
				Category:       defaultCategory,
				Username:       defaultUsername,
				SecurityIPs:    defaultSecurityIPs,
				PrivateRouting: defaultPrivateRouting,
				DatabaseName:   "testmysql",
			},
			expectedErr: nil,
		},
		{
			name: "MySQL with Customized Values",
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud":        "aws",
					"size":         20,
					"instanceType": "db.t3.micro",
					"username":     "username",
					"securityIPs": []string{
						"172.16.0.0/24",
					},
					"privateRouting": false,
				},
			},
			expected: &mysql.MySQL{
				Type:         "cloud",
				Version:      "8.0",
				Size:         20,
				InstanceType: "db.t3.micro",
				Category:     defaultCategory,
				Username:     "username",
				SecurityIPs: []string{
					"172.16.0.0/24",
				},
				PrivateRouting: false,
				DatabaseName:   "testmysql",
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		context := newGeneratorContext(project, stack, appName, workload, database,
			test.moduleInputs, tfConfigs)
		g, _ := NewMySQLGenerator(context, "testmysql", db)
		actualErr := g.(*mysqlGenerator).patchWorkspaceConfig()

		if test.expectedErr == nil {
			assert.Equal(t, test.expected, g.(*mysqlGenerator).mysql)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GetTFProviderType(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": {
			Header: database.Header{
				Type: "MySQL",
			},
			MySQL: &mysql.MySQL{
				Type:    "cloud",
				Version: "8.0",
			},
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
	db := &mysql.MySQL{
		Type:    "cloud",
		Version: "8.0",
	}

	tests := []struct {
		name         string
		moduleInputs map[string]apiv1.GenericConfig
		expected     string
		expectedErr  error
	}{
		{
			name: "AWS Provider",
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud": "aws",
				},
			},
			expected:    "aws",
			expectedErr: nil,
		},
		{
			name: "Alicloud Provider",
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {
					"cloud": "alicloud",
				},
			},
			expected:    "alicloud",
			expectedErr: nil,
		},
		{
			name:         "Empty Moudle Config Block",
			moduleInputs: map[string]apiv1.GenericConfig{},
			expected:     "",
			expectedErr:  workspace.ErrEmptyModuleConfigBlock,
		},
		{
			name: "Empty Cloud Info",
			moduleInputs: map[string]apiv1.GenericConfig{
				"mysql": {},
			},
			expected:    "",
			expectedErr: fmt.Errorf(errEmptyCloudInfo),
		},
	}

	for _, test := range tests {
		context := newGeneratorContext(project, stack, appName, workload, database,
			test.moduleInputs, tfConfigs)
		g, _ := NewMySQLGenerator(context, "testmysql", db)
		actual, actualErr := g.(*mysqlGenerator).getTFProviderType()

		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_InjectSecret(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	database := map[string]*database.Database{
		"testmysql": {
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
		"mysql": {
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
	db := &mysql.MySQL{
		Type:         "cloud",
		Version:      "8.0",
		DatabaseName: "testmysql",
	}

	data := make(map[string]string)
	data["hostAddress"] = "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
	data["username"] = "root"
	data["password"] = "$kusion_path.hashicorp:random:random_password:testapp-db.result"
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

	tests := []struct {
		name        string
		workload    *workload.Workload
		expected    container.Container
		expectedErr error
	}{
		{
			name: "Inject Secret into Service",
			workload: &workload.Workload{
				Header: workload.Header{
					Type: "Service",
				},
				Service: &workload.Service{
					Base: workload.Base{
						Containers: map[string]container.Container{
							"testcontainer": {
								Image: "testimage:latest",
							},
						},
					},
				},
			},
			expected: container.Container{
				Image: "testimage:latest",
				Env: yaml.MapSlice{
					{
						Key:   "KUSION_DB_HOST_TESTMYSQL",
						Value: "secret://" + secret.Name + "/hostAddress",
					},
					{
						Key:   "KUSION_DB_USERNAME_TESTMYSQL",
						Value: "secret://" + secret.Name + "/username",
					},
					{
						Key:   "KUSION_DB_PASSWORD_TESTMYSQL",
						Value: "secret://" + secret.Name + "/password",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Inject Secret into Job",
			workload: &workload.Workload{
				Header: workload.Header{
					Type: "Job",
				},
				Job: &workload.Job{
					Base: workload.Base{
						Containers: map[string]container.Container{
							"testcontainer": {
								Image: "testimage:latest",
							},
						},
					},
				},
			},
			expected: container.Container{
				Image: "testimage:latest",
				Env: yaml.MapSlice{
					{
						Key:   "KUSION_DB_HOST_TESTMYSQL",
						Value: "secret://" + secret.Name + "/hostAddress",
					},
					{
						Key:   "KUSION_DB_USERNAME_TESTMYSQL",
						Value: "secret://" + secret.Name + "/username",
					},
					{
						Key:   "KUSION_DB_PASSWORD_TESTMYSQL",
						Value: "secret://" + secret.Name + "/password",
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		context := newGeneratorContext(project, stack, appName, test.workload, database,
			moduleInputs, tfConfigs)
		g, _ := NewMySQLGenerator(context, "testmysql", db)
		actualErr := g.(*mysqlGenerator).injectSecret(secret)

		if test.expectedErr == nil {
			switch test.workload.Header.Type {
			case "Service":
				assert.Equal(t, test.expected, g.(*mysqlGenerator).workload.Service.Containers["testcontainer"])
				assert.NoError(t, actualErr)
			case "Job":
				assert.Equal(t, test.expected, g.(*mysqlGenerator).workload.Job.Containers["testcontainer"])
				assert.NoError(t, actualErr)
			}
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateDBSecret(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": {
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
		"mysql": {
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
		Type:         "cloud",
		Version:      "8.0",
		DatabaseName: "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name        string
		hostAddress string
		username    string
		password    string
		spec        *apiv1.Intent
		expected    *v1.Secret
		expectedErr error
	}{
		{
			name:        "Generate DB Secret",
			hostAddress: "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address",
			username:    "root",
			password:    "$kusion_path.hashicorp:random:random_password:testapp-db.result",
			spec:        &apiv1.Intent{},
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
					"hostAddress": "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address",
					"username":    "root",
					"password":    "$kusion_path.hashicorp:random:random_password:testapp-db.result",
				},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*mysqlGenerator).generateDBSecret(test.hostAddress, test.username, test.password, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateTFRandomPassword(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testmysql": {
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
		"mysql": {
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
		Type:         "cloud",
		Version:      "8.0",
		DatabaseName: "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name        string
		providerURL string
		expectedID  string
		expectedRes apiv1.Resource
	}{
		{
			name:        "Generate TF random_password",
			providerURL: "registry.terraform.io/hashicorp/random/3.5.1",
			expectedID:  "hashicorp:random:random_password:testmysql-mysql",
			expectedRes: apiv1.Resource{
				ID:   "hashicorp:random:random_password:testmysql-mysql",
				Type: "Terraform",
				Attributes: map[string]interface{}{
					"length":           16,
					"override_special": "_",
					"special":          true,
				},
				Extensions: map[string]interface{}{
					"provider":     "registry.terraform.io/hashicorp/random/3.5.1",
					"providerMeta": map[string]interface{}(nil),
					"resourceType": "random_password",
				},
			},
		},
	}

	for _, test := range tests {
		randomProvider := &inputs.Provider{}
		_ = randomProvider.SetString(test.providerURL)
		actualID, actualRes := g.(*mysqlGenerator).generateTFRandomPassword(randomProvider)

		assert.Equal(t, test.expectedID, actualID)
		assert.Equal(t, test.expectedRes, actualRes)
	}
}

func TestIsPublicAccessible(t *testing.T) {
	tests := []struct {
		name        string
		securityIPs []string
		expected    bool
	}{
		{
			name: "Public CIDR",
			securityIPs: []string{
				"0.0.0.0/0",
			},
			expected: true,
		},
		{
			name: "Private CIDR",
			securityIPs: []string{
				"172.16.0.0/24",
			},
			expected: false,
		},
		{
			name: "Private IP Address",
			securityIPs: []string{
				"172.16.0.1",
			},
		},
	}

	for _, test := range tests {
		actual := isPublicAccessible(test.securityIPs)
		assert.Equal(t, test.expected, actual)
	}
}
