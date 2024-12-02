package namespace

import (
	"testing"

	"github.com/stretchr/testify/require"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func Test_namespaceGenerator_Generate(t *testing.T) {
	type fields struct {
		namespace string
	}
	type args struct {
		Spec *v1.Spec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.Spec
		wantErr bool
	}{
		{
			name: "namespace",
			fields: fields{
				namespace: "fakeNs",
			},
			args: args{
				Spec: &v1.Spec{},
			},
			want: &v1.Spec{
				Resources: []v1.Resource{
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
			if err := g.Generate(tt.args.Spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, tt.args.Spec)
		})
	}
}
