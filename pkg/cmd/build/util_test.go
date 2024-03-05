package build

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/require"
	kclgo "kcl-lang.io/kcl-go"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/cmd/build/builders/kcl"
	"kusionstack.io/kusion/pkg/workspace"
)

var (
	intent1 = `
resources:
- id: v1:Namespace:default
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: default
      creationTimestamp: null
    spec: {}
    status: {}
`
	intentModel1 = &v1.Intent{
		Resources: []v1.Resource{
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"spec":       make(map[string]interface{}),
					"status":     make(map[string]interface{}),
					"metadata": map[string]interface{}{
						"name":              "default",
						"creationTimestamp": nil,
					},
				},
			},
		},
	}

	intent2 = `
resources:
- id: v1:Namespace:default
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: default
- id: v1:Namespace:kube-system
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: kube-system
`

	intentModel2 = &v1.Intent{
		Resources: []v1.Resource{
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "default",
					},
				},
			},
			{
				ID:   "v1:Namespace:kube-system",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "kube-system",
					},
				},
			},
		},
	}

	intentModel3 = &v1.Intent{
		Resources: []v1.Resource{
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"spec":       make(map[string]interface{}),
					"status":     make(map[string]interface{}),
					"metadata": map[string]interface{}{
						"name":              "default",
						"creationTimestamp": nil,
					},
				},
				Extensions: map[string]interface{}{
					"GVK":        "/v1, Kind=Namespace",
					"kubeConfig": "/etc/kubeconfig.yaml",
				},
			},
		},
	}

	ws = &v1.Workspace{
		Name: "default",
		Runtimes: &v1.RuntimeConfigs{
			Kubernetes: &v1.KubernetesConfig{
				KubeConfig: "/etc/kubeconfig.yaml",
			},
		},
		Backends: &v1.DeprecatedBackendConfigs{
			Local: &v1.DeprecatedLocalFileConfig{},
		},
	}
)

func TestBuildIntentFromFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    *v1.Intent
		wantErr bool
	}{
		{
			name:    "test1",
			path:    "kusion_intent.yaml",
			content: intent1,
			want:    intentModel1,
		},
		{
			name:    "test2",
			path:    "kusion_intent.yaml",
			content: intent2,
			want:    intentModel2,
		},
		{
			name:    "test3",
			path:    "kusion_intent.yaml",
			content: `k1: v1`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Create(tt.path)
			file.Write([]byte(tt.content))
			defer os.Remove(tt.path)
			got, err := IntentFromFile(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildIntent(t *testing.T) {
	apc := &v1.AppConfiguration{}
	var apcMap map[string]interface{}
	tmp, _ := json.Marshal(apc)
	_ = json.Unmarshal(tmp, &apcMap)

	type args struct {
		o       *builders.Options
		project *v1.Project
		stack   *v1.Stack
		mockers []*mockey.MockBuilder
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.Intent
		wantErr bool
	}{
		{
			name: "nil builder", args: struct {
				o       *builders.Options
				project *v1.Project
				stack   *v1.Stack
				mockers []*mockey.MockBuilder
			}{
				o: &builders.Options{Arguments: map[string]string{}},
				project: &v1.Project{
					Name: "default",
				},
				stack: &v1.Stack{
					Name: "default",
				},
				mockers: []*mockey.MockBuilder{
					mockey.Mock(kcl.Run).Return(&kcl.CompileResult{Documents: []kclgo.KCLResult{apcMap}}, nil),
					mockey.Mock(workspace.GetWorkspaceByDefaultOperator).Return(ws, nil),
				},
			},
			want: intentModel3,
		},
		{
			name: "kcl builder", args: struct {
				o       *builders.Options
				project *v1.Project
				stack   *v1.Stack
				mockers []*mockey.MockBuilder
			}{
				o: &builders.Options{},
				project: &v1.Project{
					Generator: &v1.GeneratorConfig{
						Type: v1.KCLBuilder,
					},
				},
				stack: &v1.Stack{},
				mockers: []*mockey.MockBuilder{
					mockey.Mock((*kcl.Builder).Build).Return(nil, nil),
				},
			},
			want: nil,
		},
		{
			name: "app builder", args: struct {
				o       *builders.Options
				project *v1.Project
				stack   *v1.Stack
				mockers []*mockey.MockBuilder
			}{
				o: &builders.Options{Arguments: map[string]string{}},
				project: &v1.Project{
					Name: "default",
					Generator: &v1.GeneratorConfig{
						Type: v1.AppConfigurationBuilder,
					},
				},
				stack: &v1.Stack{
					Name: "default",
				},
				mockers: []*mockey.MockBuilder{
					mockey.Mock(kcl.Run).Return(&kcl.CompileResult{Documents: []kclgo.KCLResult{apcMap}}, nil),
					mockey.Mock(workspace.GetWorkspaceByDefaultOperator).Return(ws, nil),
				},
			},
			want: intentModel3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mList []*mockey.Mocker
			for _, mocker := range tt.args.mockers {
				m := mocker.Build()
				mList = append(mList, m)
			}
			defer func() {
				for _, m := range mList {
					m.UnPatch()
				}
			}()

			got, err := Intent(tt.args.o, tt.args.project, tt.args.stack)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() got = %v, want %v", got, tt.want)
			}
		})
	}
}
