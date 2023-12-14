package init

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/apis/intent"
)

func TestValidResources(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		resources intent.Resources
	}{
		{
			name:    "valid resources",
			success: true,
			resources: []intent.Resource{
				{
					ID:   "mock-id",
					Type: "Kubernetes",
					Attributes: map[string]any{
						"mock-key": "mock-value",
					},
					Extensions: map[string]any{
						"kubeConfig": "/etc/kubeConfig.yaml",
					},
				},
			},
		},
		{
			name:    "invalid resources empty type",
			success: false,
			resources: []intent.Resource{
				{
					ID:   "mock-id",
					Type: "",
					Attributes: map[string]any{
						"mock-key": "mock-value",
					},
					Extensions: map[string]any{
						"kubeConfig": "/etc/kubeConfig.yaml",
					},
				},
			},
		},
		{
			name:    "invalid resources unsupported type",
			success: false,
			resources: []intent.Resource{
				{
					ID:   "mock-id",
					Type: "Unsupported",
					Attributes: map[string]any{
						"mock-key": "mock-value",
					},
					Extensions: map[string]any{
						"kubeConfig": "/etc/kubeConfig.yaml",
					},
				},
			},
		},
		{
			name:    "invalid resources multiple kubeConfig",
			success: false,
			resources: []intent.Resource{
				{
					ID:   "mock-id",
					Type: "Kubernetes",
					Attributes: map[string]any{
						"mock-key": "mock-value",
					},
					Extensions: map[string]any{
						"kubeConfig": "/etc/kubeConfig.yaml",
					},
				},
				{
					ID:   "mock-id",
					Type: "Kubernetes",
					Attributes: map[string]any{
						"mock-key": "mock-value",
					},
					Extensions: map[string]any{
						"kubeConfig": "/etc/kubeConfig_2.yaml",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validResources(tc.resources)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}
