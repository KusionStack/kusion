package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
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

func TestNewDatabaseGenerator(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)

	tests := []struct {
		name        string
		data        modules.GeneratorContext
		expected    *databaseGenerator
		expectedErr error
	}{
		{
			name: "Valid Database Generator",
			data: context,
			expected: &databaseGenerator{
				project:       project,
				stack:         stack,
				appName:       appName,
				workload:      workload,
				database:      database,
				moduleConfigs: moduleInputs,
				tfConfigs:     tfConfigs,
				namespace:     project.Name,
				context:       context,
			},
		},
	}

	for _, test := range tests {
		actual, actualErr := NewDatabaseGenerator(test.data)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestDatabaseGenerator_NewDatabaseGeneratorFunc(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)

	tests := []struct {
		name        string
		data        modules.GeneratorContext
		expected    modules.Generator
		expectedErr error
	}{
		{
			name: "Valid Database Generator Func",
			data: context,
			expected: &databaseGenerator{
				project:       project,
				stack:         stack,
				appName:       appName,
				workload:      workload,
				database:      database,
				moduleConfigs: moduleInputs,
				tfConfigs:     tfConfigs,
				namespace:     project.Name,
				context:       context,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		testGeneratorFunc := NewDatabaseGeneratorFunc(test.data)
		actual, actualErr := testGeneratorFunc()
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestGenerate(t *testing.T) {
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
		expectedErr  error
	}{
		{
			name: "Generate Local MySQL Database",
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
			expectedErr:  nil,
		},
		{
			name: "Generate Local PostgreSQL Database",
			database: map[string]*database.Database{
				"testpostgres": {
					Header: database.Header{
						Type: "PostgreSQL",
					},
					PostgreSQL: &postgres.PostgreSQL{
						Type:    "local",
						Version: "14.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{},
			expectedErr:  nil,
		},
		{
			name: "Generate Local MySQL And PostgreSQL Database",
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
				"testpostgres": {
					Header: database.Header{
						Type: "PostgreSQL",
					},
					PostgreSQL: &postgres.PostgreSQL{
						Type:    "local",
						Version: "14.0",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{},
			expectedErr:  nil,
		},
		{
			name: "Generate Unsupported Database Type",
			database: map[string]*database.Database{
				"test": {
					Header: database.Header{
						Type: "Unsupported",
					},
				},
			},
			moduleInputs: map[string]apiv1.GenericConfig{},
			expectedErr:  fmt.Errorf(errUnsupportedDatabaseType, "Unsupported"),
		},
	}

	for _, test := range tests {
		context := newGeneratorContext(project, stack,
			appName, workload, test.database, test.moduleInputs, tfConfigs)
		g, _ := NewDatabaseGenerator(context)
		actualErr := g.(*databaseGenerator).Generate(&apiv1.Intent{})

		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}
