package builders

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kcl-lang.io/kpm/pkg/api"
	pkg "kcl-lang.io/kpm/pkg/package"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	internalv1 "kusionstack.io/kusion/pkg/apis/internal.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
)

func TestAppsConfigBuilder_Build(t *testing.T) {
	p, s := buildMockProjectAndStack()
	appName, app := buildMockApp()
	acg := &AppsConfigBuilder{
		Apps: map[string]internalv1.AppConfiguration{
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

func buildMockApp() (string, *internalv1.AppConfiguration) {
	return "app1", &internalv1.AppConfiguration{
		Workload: &internalv1.Workload{
			Header: internalv1.Header{
				Type: "Service",
			},
			Service: &internalv1.Service{
				Base: internalv1.Base{},
				Type: "Deployment",
				Ports: []internalv1.Port{
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
