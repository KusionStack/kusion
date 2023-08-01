package spec

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"kusionstack.io/kusion/pkg/models"
)

var (
	spec1 = `
- id: apps/v1:Namespace:default
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: default
`
	specModle1 = &models.Spec{
		Resources: []models.Resource{
			{
				ID:   "apps/v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "default",
					},
				},
			},
		},
	}

	spec2 = `
- id: apps/v1:Namespace:default
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: default
- id: apps/v1:Namespace:kube-system
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: kube-system
`

	specModle2 = &models.Spec{
		Resources: []models.Resource{
			{
				ID:   "apps/v1:Namespace:default",
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
				ID:   "apps/v1:Namespace:kube-system",
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
)

func TestGenerateSpecFromFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    *models.Spec
		wantErr bool
	}{
		{
			name:    "test1",
			path:    "kusion_spec.yaml",
			content: spec1,
			want:    specModle1,
		},
		{
			name:    "test2",
			path:    "kusion_spec.yaml",
			content: spec2,
			want:    specModle2,
		},
		{
			name:    "test3",
			path:    "kusion_spec.yaml",
			content: `k1: v1`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Create(tt.path)
			file.Write([]byte(tt.content))
			defer os.Remove(tt.path)
			got, err := GenerateSpecFromFile(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
