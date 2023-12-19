package trait

import (
	"testing"

	"github.com/stretchr/testify/require"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	appmodule "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func Test_opsRuleGenerator_Generate(t *testing.T) {
	type fields struct {
		project *project.Project
		stack   *stack.Stack
		appName string
		app     *appmodule.AppConfiguration
	}
	type args struct {
		spec *intent.Intent
	}
	project := &project.Project{
		Configuration: project.Configuration{
			Name: "default",
		},
	}
	stack := &stack.Stack{
		Configuration: stack.Configuration{Name: "dev"},
	}
	appName := "foo"
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		exp     *intent.Intent
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
				spec: &intent.Intent{},
			},
			wantErr: false,
			exp:     &intent.Intent{},
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
				spec: &intent.Intent{},
			},
			wantErr: false,
			exp: &intent.Intent{
				Resources: intent.Resources{
					intent.Resource{
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
		project *project.Project
		stack   *stack.Stack
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
