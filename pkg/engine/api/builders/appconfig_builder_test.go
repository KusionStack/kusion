package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload/network"

	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
)

func TestAppsConfigBuilder_Build(t *testing.T) {
	p, s := buildMockProjectAndStack()
	appName, app := buildMockApp()
	acg := &AppsConfigBuilder{
		Apps: map[string]v1.AppConfiguration{
			appName: *app,
		},
		Workspace: buildMockWorkspace(),
	}

	intent, err := acg.Build(&Options{}, p, s)
	assert.NoError(t, err)
	assert.NotNil(t, intent)
}

func buildMockApp() (string, *v1.AppConfiguration) {
	return "app1", &v1.AppConfiguration{
		Workload: &workload.Workload{
			Header: workload.Header{
				Type: "Service",
			},
			Service: &workload.Service{
				Base: workload.Base{},
				Type: "Deployment",
				Ports: []network.Port{
					{
						Port:     80,
						Protocol: "TCP",
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

func buildMockProjectAndStack() (*v1.Project, *v1.Stack) {
	p := &v1.Project{
		Name: "test-project",
	}

	s := &v1.Stack{
		Name: "test-project",
	}

	return p, s
}
