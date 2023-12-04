package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/apis/workspace"
	appmodel "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/network"
)

func TestAppsConfigBuilder_Build(t *testing.T) {
	p, s := buildMockProjectAndStack()
	appName, app := buildMockApp()
	acg := &AppsConfigBuilder{
		Apps: map[string]appmodel.AppConfiguration{
			appName: *app,
		},
		Workspace: buildMockWorkspace(),
	}

	intent, err := acg.Build(&Options{}, p, s)
	assert.NoError(t, err)
	assert.NotNil(t, intent)
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

func buildMockWorkspace() *workspace.Workspace {
	return &workspace.Workspace{
		Name: "test",
		Modules: workspace.ModuleConfigs{
			"database": {
				"default": {
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.micro",
				},
				"smallClass": {
					"instanceType":    "db.t3.small",
					"projectSelector": []any{"foo", "bar"},
				},
			},
			"port": {
				"default": {
					"type": "aws",
				},
			},
		},
		Runtimes: &workspace.RuntimeConfigs{
			Kubernetes: &workspace.KubernetesConfig{
				KubeConfig: "/etc/kubeconfig.yaml",
			},
		},
		Backends: &workspace.BackendConfigs{
			Local: &workspace.LocalFileConfig{
				Path: "/etc/.kusion",
			},
		},
	}
}

func buildMockProjectAndStack() (*project.Project, *stack.Stack) {
	p := &project.Project{
		Configuration: project.Configuration{
			Name: "test-project",
		},
	}

	s := &stack.Stack{
		Configuration: stack.Configuration{
			Name: "test-stack",
		},
	}

	return p, s
}
