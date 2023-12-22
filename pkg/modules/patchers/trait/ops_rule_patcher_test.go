package trait

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	modelsapp "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
)

func Test_opsRulePatcher_Patch(t *testing.T) {
	i := &apiv1.Intent{}
	err := modules.AppendToIntent(apiv1.Kubernetes, "id", i, buildMockDeployment())
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		app             *modelsapp.AppConfiguration
		workspaceConfig map[string]apiv1.GenericConfig
	}
	type args struct {
		resources map[string][]*apiv1.Resource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Patch Deployment",
			fields: fields{
				app: &modelsapp.AppConfiguration{
					OpsRule: &trait.OpsRule{
						MaxUnavailable: "30%",
					},
				},
			},
			args: args{
				resources: i.Resources.GVKIndex(),
			},
		},
		{
			name: "Patch Deployment with workspace config",
			fields: fields{
				app: &modelsapp.AppConfiguration{},
				workspaceConfig: map[string]apiv1.GenericConfig{
					"opsRule": {
						"maxUnavailable": "30%",
					},
				},
			},
			args: args{
				resources: i.Resources.GVKIndex(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &opsRulePatcher{
				app:           tt.fields.app,
				modulesConfig: tt.fields.workspaceConfig,
			}
			if err := p.Patch(tt.args.resources); (err != nil) != tt.wantErr {
				t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
			}
			// check if the deployment is patched
			var deployment appsv1.Deployment
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(i.Resources[0].Attributes, &deployment); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, appsv1.RollingUpdateDeploymentStrategyType, deployment.Spec.Strategy.Type)
			assert.NotNil(t, deployment.Spec.Strategy.RollingUpdate)
			if tt.fields.app.OpsRule != nil {
				assert.Equal(t, intstr.Parse(tt.fields.app.OpsRule.MaxUnavailable), *deployment.Spec.Strategy.RollingUpdate.MaxUnavailable)
			} else {
				assert.Equal(t, tt.fields.workspaceConfig["opsRule"]["maxUnavailable"],
					(*deployment.Spec.Strategy.RollingUpdate.MaxUnavailable).String())
			}
		})
	}
}

// generate a mock Deployment
func buildMockDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mock-deployment",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		Spec: appsv1.DeploymentSpec{},
	}
}

func TestNewOpsRulePatcherFunc(t *testing.T) {
	p := &apiv1.Project{
		Name: "default",
	}
	type args struct {
		app       *modelsapp.AppConfiguration
		project   *apiv1.Project
		workspace map[string]apiv1.GenericConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "NewOpsRulePatcherFunc",
			args: args{
				app: &modelsapp.AppConfiguration{},
				workspace: map[string]apiv1.GenericConfig{
					"opsRule": {
						"maxUnavailable": "30%",
					},
				},
				project: p,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcherFunc := NewOpsRulePatcherFunc(tt.args.app, tt.args.workspace)
			assert.NotNil(t, patcherFunc)
			patcher, err := patcherFunc()
			assert.NoError(t, err)
			assert.Equal(t, tt.args.app, patcher.(*opsRulePatcher).app)
		})
	}
}
