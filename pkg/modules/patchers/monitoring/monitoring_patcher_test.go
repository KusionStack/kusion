package monitoring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	modelsapp "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/monitoring"
)

func Test_monitoringPatcher_Patch(t *testing.T) {
	i := &apiv1.Intent{}
	err := modules.AppendToIntent(apiv1.Kubernetes, "id", i, buildMockDeployment())
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		app       *modelsapp.AppConfiguration
		workspace map[string]apiv1.GenericConfig
	}
	type args struct {
		resources map[string][]*apiv1.Resource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "operatorModeTrue",
			fields: fields{
				app: &modelsapp.AppConfiguration{
					Name: "test-app",
					Monitoring: &monitoring.Monitor{
						Path: "/metrics",
						Port: "web",
					},
				},
				workspace: map[string]apiv1.GenericConfig{
					"monitoring": {
						"operatorMode": true,
						"monitorType":  "Pod",
						"scheme":       "http",
						"interval":     "30s",
						"timeout":      "15s",
					},
				},
			},
			args: args{
				resources: i.Resources.GVKIndex(),
			},
			wantErr: assert.NoError,
		},
		{
			name: "operatorModeFalse",
			fields: fields{
				app: &modelsapp.AppConfiguration{
					Name: "test-app",
					Monitoring: &monitoring.Monitor{
						Path: "/metrics",
						Port: "8080",
					},
				},
				workspace: map[string]apiv1.GenericConfig{
					"monitoring": {
						"operatorMode": false,
						"scheme":       "http",
						"interval":     "30s",
						"timeout":      "15s",
					},
				},
			},
			args: args{
				resources: i.Resources.GVKIndex(),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &monitoringPatcher{
				app:           tt.fields.app,
				modulesConfig: tt.fields.workspace,
			}
			tt.wantErr(t, p.Patch(tt.args.resources), fmt.Sprintf("Patch(%v)", tt.args.resources))
			// check if the deployment is patched
			var deployment appsv1.Deployment
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(i.Resources[0].Attributes, &deployment); err != nil {
				t.Fatal(err)
			}
			if tt.fields.app.Monitoring.OperatorMode {
				assert.NotNil(t, deployment.Labels)
				assert.NotNil(t, deployment.Spec.Template.Labels)
				assert.Equal(t, deployment.Labels["kusion_monitoring_appname"], tt.fields.app.Name)
				assert.Equal(t, deployment.Spec.Template.Labels["kusion_monitoring_appname"], tt.fields.app.Name)
			} else {
				assert.NotNil(t, deployment.Annotations)
				assert.NotNil(t, deployment.Spec.Template.Annotations)
				assert.Equal(t, deployment.Annotations["prometheus.io/scrape"], "true")
				assert.Equal(t, deployment.Annotations["prometheus.io/scheme"], tt.fields.app.Monitoring.Scheme)
				assert.Equal(t, deployment.Annotations["prometheus.io/path"], tt.fields.app.Monitoring.Path)
				assert.Equal(t, deployment.Annotations["prometheus.io/port"], tt.fields.app.Monitoring.Port)
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/scrape"], "true")
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/scheme"], tt.fields.app.Monitoring.Scheme)
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/path"], tt.fields.app.Monitoring.Path)
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/port"], tt.fields.app.Monitoring.Port)
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

func TestNewMonitoringPatcherFunc(t *testing.T) {
	type args struct {
		app       *modelsapp.AppConfiguration
		workspace map[string]apiv1.GenericConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "NewMonitoringPatcherFunc",
			args: args{
				app: &modelsapp.AppConfiguration{
					Name: "test-app",
					Monitoring: &monitoring.Monitor{
						Path: "/metrics",
						Port: "web",
					},
				},
				workspace: map[string]apiv1.GenericConfig{
					"monitoring": {
						"operatorMode": true,
						"monitorType":  "Pod",
						"scheme":       "http",
						"interval":     "15s",
						"timeout":      "30s",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcherFunc := NewMonitoringPatcherFunc(tt.args.app, tt.args.workspace)
			assert.NotNil(t, patcherFunc)
			patcher, err := patcherFunc()
			assert.NoError(t, err)
			assert.Equal(t, tt.args.app.Name, patcher.(*monitoringPatcher).app.Name)
		})
	}
}
