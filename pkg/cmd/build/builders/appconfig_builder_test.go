package builders

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kcl-lang.io/kpm/pkg/api"
	pkg "kcl-lang.io/kpm/pkg/package"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload/network"
	"kusionstack.io/kusion/pkg/modules"

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

	kpmMock := mockey.Mock((*api.KclPackage).GetDependenciesInModFile).Return(&pkg.Dependencies{Deps: make(map[string]pkg.Dependency)}).
		Build()
	callMock := mockey.Mock(modules.CallGenerators).Return(nil).Build()
	defer func() {
		kpmMock.UnPatch()
		callMock.UnPatch()
	}()

	kclPkg := &api.KclPackage{}
	intent, err := acg.Build(kclPkg, p, s)
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
		Runtimes: &v1.RuntimeConfigs{
			Kubernetes: &v1.KubernetesConfig{
				KubeConfig: "/etc/kubeconfig.yaml",
			},
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
