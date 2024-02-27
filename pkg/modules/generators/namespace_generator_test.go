package generators

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func Test_namespaceGenerator_Generate(t *testing.T) {
	type fields struct {
		namespace string
	}
	type args struct {
		intent *apiv1.Intent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apiv1.Intent
		wantErr bool
	}{
		{
			name: "namespace",
			fields: fields{
				namespace: "fakeNs",
			},
			args: args{
				intent: &apiv1.Intent{},
			},
			want: &apiv1.Intent{
				Resources: []apiv1.Resource{
					{
						ID:   "v1:Namespace:fakeNs",
						Type: "Kubernetes",
						Attributes: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Namespace",
							"metadata": map[string]interface{}{
								"creationTimestamp": nil,
								"name":              "fakeNs",
							},
							"spec":   make(map[string]interface{}),
							"status": make(map[string]interface{}),
						},
						DependsOn: nil,
						Extensions: map[string]interface{}{
							"GVK": "/v1, Kind=Namespace",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &namespaceGenerator{
				namespace: tt.fields.namespace,
			}
			if err := g.Generate(tt.args.intent); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, tt.args.intent)
		})
	}
}
