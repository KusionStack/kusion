package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	appmodel "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/network"
)

func TestAppConfigurationGenerator_Generate(t *testing.T) {
	project, stack := buildMockProjectAndStack()
	appName, app := buildMockApp()

	g := &appConfigurationGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		app:     app,
	}

	spec := &intent.Intent{
		Resources: []intent.Resource{},
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
		assert.EqualError(t, err, "can not find app configuration when generating the Intent")
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
				Ports: []network.Port{
					{
						Type:     network.CSPAliyun,
						Port:     80,
						Protocol: "TCP",
						Public:   true,
					},
				},
			},
		},
	}
}

func buildMockProjectAndStack() (*project.Project, *stack.Stack) {
	project := &project.Project{
		ProjectConfiguration: project.ProjectConfiguration{
			Name: "testproject",
		},
	}

	stack := &stack.Stack{
		Configuration: stack.Configuration{
			Name: "teststack",
		},
	}

	return project, stack
}
