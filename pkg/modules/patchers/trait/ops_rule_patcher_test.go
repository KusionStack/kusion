package trait

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules"
	modelsapp "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
)

func Test_opsRulePatcher_Patch(t *testing.T) {
	i := &intent.Intent{}
	err := modules.AppendToIntent(intent.Kubernetes, "id", i, buildMockDeployment())
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		app *modelsapp.AppConfiguration
	}
	type args struct {
		resources map[string][]*intent.Resource
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &opsRulePatcher{
				app: tt.fields.app,
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
			assert.Equal(t, intstr.Parse(tt.fields.app.OpsRule.MaxUnavailable), *deployment.Spec.Strategy.RollingUpdate.MaxUnavailable)
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
	type args struct {
		app *modelsapp.AppConfiguration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "NewOpsRulePatcherFunc",
			args: args{
				app: &modelsapp.AppConfiguration{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcherFunc := NewOpsRulePatcherFunc(tt.args.app)
			assert.NotNil(t, patcherFunc)
			patcher, err := patcherFunc()
			assert.NoError(t, err)
			assert.Equal(t, tt.args.app, patcher.(*opsRulePatcher).app)
		})
	}
}
