package generators

import (
	"testing"

	"github.com/stretchr/testify/require"

	"kusionstack.io/kusion/pkg/apis/intent"
)

func Test_namespaceGenerator_Generate(t *testing.T) {
	type fields struct {
		projectName string
	}
	type args struct {
		spec *intent.Intent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *intent.Intent
		wantErr bool
	}{
		{
			name: "namespace",
			fields: fields{
				projectName: "fake-project",
			},
			args: args{
				spec: &intent.Intent{},
			},
			want: &intent.Intent{
				Resources: []intent.Resource{
					{
						ID:   "v1:Namespace:fake-project",
						Type: "Kubernetes",
						Attributes: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Namespace",
							"metadata": map[string]interface{}{
								"creationTimestamp": nil,
								"name":              "fake-project",
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
				projectName: tt.fields.projectName,
			}
			if err := g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, tt.args.spec)
		})
	}
}
