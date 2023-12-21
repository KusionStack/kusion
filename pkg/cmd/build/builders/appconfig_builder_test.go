package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
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

func buildMockWorkspace() *v1.Workspace {
	return &v1.Workspace{
		Name: "test",
		Modules: v1.ModuleConfigs{
			"database": {
				Default: v1.GenericConfig{
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.micro",
				},
				ModulePatcherConfigs: v1.ModulePatcherConfigs{
					"smallClass": {
						GenericConfig: v1.GenericConfig{
							"instanceType": "db.t3.small",
						},
						ProjectSelector: []string{"foo", "bar"},
					},
				},
			},
			"port": {
				Default: v1.GenericConfig{
					"type": "aws",
				},
			},
		},
		Runtimes: &v1.RuntimeConfigs{
			Kubernetes: &v1.KubernetesConfig{
				KubeConfig: "/etc/kubeconfig.yaml",
			},
		},
		Backends: &v1.BackendConfigs{
			Local: &v1.LocalFileConfig{},
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
