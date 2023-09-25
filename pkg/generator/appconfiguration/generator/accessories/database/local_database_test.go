package accessories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/accessories/database"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

func TestGenerateLocalResources(t *testing.T) {
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "testproject",
		},
	}
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:     "local",
		Engine:   "MariaDB",
		Version:  "10.5",
		Size:     10,
		Username: "root",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	spec := &models.Spec{}
	secret, err := generator.generateLocalResources(database, spec)

	hostAddress := "testapp-db-local-service"
	username := database.Username
	password := "$kusion_path.v1:Secret:testproject:testapp-db-local-secret.stringData.password"
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

func TestGenerateLocalSecret(t *testing.T) {
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "testproject",
		},
	}
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:     "local",
		Engine:   "MariaDB",
		Version:  "10.5",
		Size:     10,
		Username: "root",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	spec := &models.Spec{}
	kusionPathSecret, err := generator.generateLocalSecret(spec)
	expectedKusionPathSec := "$kusion_path.v1:Secret:testproject:testapp-db-local-secret.stringData.password"

	assert.NoError(t, err)
	assert.Equal(t, expectedKusionPathSec, kusionPathSecret)
}

func TestGenerateLocalPVC(t *testing.T) {
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "testproject",
		},
	}
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:     "local",
		Engine:   "MariaDB",
		Version:  "10.5",
		Size:     10,
		Username: "root",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	spec := &models.Spec{}
	err := generator.generateLocalPVC(database, spec)

	assert.NoError(t, err)
}

func TestGenerateLocalDeployment(t *testing.T) {
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "testproject",
		},
	}
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:     "local",
		Engine:   "MariaDB",
		Version:  "10.5",
		Size:     10,
		Username: "root",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	spec := &models.Spec{}
	err := generator.generateLocalDeployment(database, spec)

	assert.NoError(t, err)
}

func TestGenerateLocalService(t *testing.T) {
	project := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "testproject",
		},
	}
	stack := &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "teststack",
		},
	}
	appName := "testapp"
	workload := &workload.Workload{}
	database := &database.Database{
		Type:     "local",
		Engine:   "MariaDB",
		Version:  "10.5",
		Size:     10,
		Username: "root",
	}
	generator := &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
	}

	spec := &models.Spec{}
	svcName, err := generator.generateLocalService(database, spec)
	expectedSvcName := "testapp-db-local-service"

	assert.NoError(t, err)
	assert.Equal(t, expectedSvcName, svcName)
}
