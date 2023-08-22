package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	appmodel "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

func TestAppsGenerator_GenerateSpec(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()
	acg := &AppsGenerator{
		Apps: map[string]appmodel.AppConfiguration{
			appName: *app,
		},
	}

	spec, err := acg.GenerateSpec(&generator.Options{}, project, stack)
	assert.NoError(t, err)
	assert.NotNil(t, spec)
}

func TestAppConfigurationGenerator_Generate(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()

	g := &appConfigurationGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
	}

	spec := &models.Spec{
		Resources: []models.Resource{},
	}

	err := g.Generate(spec)
	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Resources)
}

func TestNewAppConfigurationGeneratorFunc(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()

	t.Run("Valid app configuration generator func", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, app)()
		assert.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Empty app name", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, "", app)()
		assert.EqualError(t, err, "app name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Nil app", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, nil)()
		assert.EqualError(t, err, "can not find app configuration when generating the Spec")
		assert.Nil(t, g)
	})

	t.Run("Empty project name", func(t *testing.T) {
		project.Name = ""
		g, err := NewAppConfigurationGeneratorFunc(project, stack, appName, nil)()
		assert.EqualError(t, err, "project name must not be empty")
		assert.Nil(t, g)
	})
}

func buildMockApp() (string, *appmodel.AppConfiguration) {
	return "app1", &appmodel.AppConfiguration{
		Workload: &workload.Workload{
			Header: workload.Header{
				Type: "Service",
			},
			Service: &workload.Service{
				Base: workload.Base{},
				Type: "Deployment",
			},
		},
	}
}

func buildMockProjectAndStack() (*projectstack.Project, *projectstack.Stack) {
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

	return project, stack
}
