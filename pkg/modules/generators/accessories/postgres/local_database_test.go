package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestPostgresGenerator_GenerateLocalResources(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": &database.Database{
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testpostgres",
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
			name: "Generate Local Resources",
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
					"hostAddress": "testpostgres-postgres-local-service",
					"username":    "root",
					"password":    g.(*postgresGenerator).generateLocalPassword(16),
				},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*postgresGenerator).generateLocalResources(test.db, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestPostgresGenerator_GenerateLocalSecret(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": &database.Database{
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name        string
		spec        *apiv1.Intent
		expected    string
		expectedErr error
	}{
		{
			name:        "Generate Local Secret",
			spec:        &apiv1.Intent{},
			expected:    g.(*postgresGenerator).generateLocalPassword(16),
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*postgresGenerator).generateLocalSecret(test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestPostgresGenerator_GenerateLocalPVC(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": &database.Database{
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name        string
		db          *postgres.PostgreSQL
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
		actualErr := g.(*postgresGenerator).generateLocalPVC(test.db, test.spec)
		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestPostgresGenerator_GenerateLocalDeployment(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": &database.Database{
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name        string
		db          *postgres.PostgreSQL
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
		actualErr := g.(*postgresGenerator).generateLocalDeployment(test.db, test.spec)
		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}

func TestPostgresGenerator_GenerateLocalService(t *testing.T) {
	project := &apiv1.Project{Name: "testproject"}
	stack := &apiv1.Stack{Name: "teststack"}
	appName := "testapp"
	workload := &workload.Workload{}
	database := map[string]*database.Database{
		"testpostgres": &database.Database{
			Header: database.Header{
				Type: "PostgreSQL",
			},
			PostgreSQL: &postgres.PostgreSQL{
				Type:    "local",
				Version: "8.0",
			},
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{}
	tfConfigs := apiv1.TerraformConfig{}
	context := newGeneratorContext(project, stack, appName, workload, database,
		moduleInputs, tfConfigs)
	db := &postgres.PostgreSQL{
		Type:         "local",
		Version:      "8.0",
		Username:     "root",
		DatabaseName: "testpostgres",
	}
	g, _ := NewPostgresGenerator(context, "testpostgres", db)

	tests := []struct {
		name        string
		db          *postgres.PostgreSQL
		spec        *apiv1.Intent
		expected    string
		expectedErr error
	}{
		{
			name:        "Generate Local Service",
			db:          db,
			spec:        &apiv1.Intent{},
			expected:    "testpostgres-postgres-local-service",
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		actual, actualErr := g.(*postgresGenerator).generateLocalService(test.db, test.spec)
		if test.expectedErr == nil {
			assert.Equal(t, test.expected, actual)
			assert.NoError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, test.expectedErr.Error())
		}
	}
}
