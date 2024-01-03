package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestMySQLGenerator_GenerateLocalResources(t *testing.T) {
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
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testmysql",
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
			name: "Generate Local Resources",
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
					"hostAddress": "testmysql-mysql-local-service",
					"username":    "root",
					"password":    g.(*mysqlGenerator).generateLocalPassword(16),
				},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*mysqlGenerator).generateLocalResources(test.db, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateLocalSecret(t *testing.T) {
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
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name        string
		spec        *apiv1.Intent
		expected    string
		expectedErr error
	}{
		{
			name:        "Generate Local Secret",
			spec:        &apiv1.Intent{},
			expected:    g.(*mysqlGenerator).generateLocalPassword(16),
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*mysqlGenerator).generateLocalSecret(test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateLocalPVC(t *testing.T) {
	g, db := newDefaultMySQLGenerator()

	tests := []struct {
		name        string
		db          *mysql.MySQL
		spec        *apiv1.Intent
		expectedErr error
	}{
		{
			name:        "Generate Local PVC",
			db:          db,
			spec:        &apiv1.Intent{},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actualErr := g.generateLocalPVC(test.db, test.spec)
		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateLocalDeployment(t *testing.T) {
	g, db := newDefaultMySQLGenerator()

	tests := []struct {
		name        string
		db          *mysql.MySQL
		spec        *apiv1.Intent
		expectedErr error
	}{
		{
			name:        "Generate Local Deployment",
			db:          db,
			spec:        &apiv1.Intent{},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actualErr := g.generateLocalDeployment(test.db, test.spec)
		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestMySQLGenerator_GenerateLocalService(t *testing.T) {
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
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	tests := []struct {
		name        string
		db          *mysql.MySQL
		spec        *apiv1.Intent
		expected    string
		expectedErr error
	}{
		{
			name:        "Generate Local Service",
			db:          db,
			spec:        &apiv1.Intent{},
			expected:    "testmysql-mysql-local-service",
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*mysqlGenerator).generateLocalService(test.db, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func newDefaultMySQLGenerator() (*mysqlGenerator, *mysql.MySQL) {
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
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &mysql.MySQL{
		Type:         "local",
		Version:      "8.0",
		DatabaseName: "testmysql",
	}
	g, _ := NewMySQLGenerator(context, "testmysql", db)

	return g.(*mysqlGenerator), db
}
