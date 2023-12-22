package trait

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	appmodule "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func Test_opsRuleGenerator_Generate(t *testing.T) {
	type fields struct {
		project *apiv1.Project
		stack   *apiv1.Stack
		appName string
		app     *appmodule.AppConfiguration
	}
	type args struct {
		spec *apiv1.Intent
	}
	project := &apiv1.Project{
		Name: "default",
	}
	stack := &apiv1.Stack{
		Name: "dev",
	}
	appName := "foo"
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		exp     *apiv1.Intent
	}{
		{
			name: "test Job",
			fields: fields{
				project: project,
				stack:   stack,
				appName: appName,
				app: &appmodule.AppConfiguration{
					Workload: &workload.Workload{
						Header: workload.Header{
							Type: workload.TypeJob,
						},
					},
					OpsRule: &trait.OpsRule{
						MaxUnavailable: "30%",
					},
				},
			},
			args: args{
				spec: &apiv1.Intent{},
			},
			wantErr: false,
			exp:     &apiv1.Intent{},
		},
		{
			name: "test CollaSet",
			fields: fields{
				project: project,
				stack:   stack,
				appName: appName,
				app: &appmodule.AppConfiguration{
					Workload: &workload.Workload{
						Header: workload.Header{
							Type: workload.TypeService,
						},
						Service: &workload.Service{
							Type: workload.TypeCollaset,
						},
					},
					OpsRule: &trait.OpsRule{
						MaxUnavailable: "30%",
					},
				},
			},
			args: args{
				spec: &apiv1.Intent{},
			},
			wantErr: false,
			exp: &apiv1.Intent{
				Resources: apiv1.Resources{
					apiv1.Resource{
						ID:   "apps.kusionstack.io/v1alpha1:PodTransitionRule:default:default-dev-foo",
						Type: "Kubernetes",
						Attributes: map[string]interface{}{
							"apiVersion": "apps.kusionstack.io/v1alpha1",
							"kind":       "PodTransitionRule",
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
								"name":              "default-dev-foo",
								"namespace":         "default",
							},
							"spec": map[string]interface{}{
								"rules": []interface{}{map[string]interface{}{
									"availablePolicy": map[string]interface{}{
										"maxUnavailableValue": "30%",
									},
									"name": "maxUnavailable",
								}},
								"selector": map[string]interface{}{
									"matchLabels": map[string]interface{}{
										"app.kubernetes.io/name": "foo", "app.kubernetes.io/part-of": "default",
									},
								},
							}, "status": map[string]interface{}{},
						},
						DependsOn: []string(nil),
						Extensions: map[string]interface{}{
							"GVK": "apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &opsRuleGenerator{
				project: tt.fields.project,
				stack:   tt.fields.stack,
				appName: tt.fields.appName,
				app:     tt.fields.app,
			}
			err := g.Generate(tt.args.spec)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.exp, tt.args.spec)
			}
		})
	}
}

func TestNewOpsRuleGeneratorFunc(t *testing.T) {
	type args struct {
		project *apiv1.Project
		stack   *apiv1.Stack
		appName string
		app     *appmodule.AppConfiguration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *opsRuleGenerator
	}{
		{
			name: "test1",
			args: args{
				project: nil,
				stack:   nil,
				appName: "",
				app:     nil,
			},
			wantErr: false,
			want:    &opsRuleGenerator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewOpsRuleGeneratorFunc(tt.args.project, tt.args.stack, tt.args.appName, tt.args.app)
			g, err := f()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, g)
			}
		})
	}
}
