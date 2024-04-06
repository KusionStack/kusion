package init

import (
	"testing"

	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func TestValidResources(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		resources apiv1.Resources
	}{
		{
			name:    "valid resources",
			success: true,
			resources: []apiv1.Resource{
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
			resources: []apiv1.Resource{
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
			resources: []apiv1.Resource{
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
			resources: []apiv1.Resource{
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
