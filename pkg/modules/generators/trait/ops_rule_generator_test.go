package trait

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	appmodule "kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/trait"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func Test_opsRuleGenerator_Generate(t *testing.T) {
	type fields struct {
		project         *apiv1.Project
		stack           *apiv1.Stack
		appName         string
		app             *appmodule.AppConfiguration
		workspaceConfig map[string]apiv1.GenericConfig
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
			name: "test CollaSet with opsRule in AppConfig",
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
		{
			name: "test CollaSet with opsRule in workspace",
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
				},
				workspaceConfig: map[string]apiv1.GenericConfig{
					"opsRule": {
						"maxUnavailable": 7,
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
										"maxUnavailableValue": 7,
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
				project:       tt.fields.project,
				stack:         tt.fields.stack,
				appName:       tt.fields.appName,
				app:           tt.fields.app,
				modulesConfig: tt.fields.workspaceConfig,
				namespace:     tt.fields.project.Name,
			}
			err := g.Generate(tt.args.spec)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				exp, _ := json.Marshal(tt.exp)
				act, _ := json.Marshal(tt.args.spec)
				require.Equal(t, exp, act)
			}
		})
	}
}

func TestNewOpsRuleGeneratorFunc(t *testing.T) {
	p := &apiv1.Project{
		Name: "default",
	}
	s := &apiv1.Stack{
		Name: "dev",
	}
	app := &appmodule.AppConfiguration{
		Name: "beep",
	}

	type args struct {
		project *apiv1.Project
		stack   *apiv1.Stack
		appName string
		app     *appmodule.AppConfiguration
		ws      map[string]apiv1.GenericConfig
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
				project: p,
				stack:   s,
				appName: "",
				app:     app,
				ws: map[string]apiv1.GenericConfig{
					"opsRule": {
						"maxUnavailable": "30%",
					},
				},
			},
			wantErr: false,
			want: &opsRuleGenerator{
				project: p,
				stack:   s,
				appName: "beep",
				app:     app,
				modulesConfig: map[string]apiv1.GenericConfig{
					"opsRule": {
						"maxUnavailable": "30%",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := modules.GeneratorContext{
				Project:      tt.args.project,
				Stack:        tt.args.stack,
				Application:  tt.args.app,
				Namespace:    tt.args.project.Name,
				ModuleInputs: tt.args.ws,
			}
			f := NewOpsRuleGeneratorFunc(context)
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
