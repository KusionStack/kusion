package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func TestPortForwardOperation_PortForward(t *testing.T) {
	testcases := []struct {
		name        string
		req         *PortForwardRequest
		expectedErr bool
	}{
		{
			name: "empty spec",
			req: &PortForwardRequest{
				Port: 8080,
			},
			expectedErr: true,
		},
		{
			name: "empty services",
			req: &PortForwardRequest{
				Spec: &v1.Spec{
					Resources: v1.Resources{
						{
							ID:   "v1:Namespace:quickstart",
							Type: "Kubernetes",
							Attributes: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Namespace",
								"metadata": map[string]interface{}{
									"name": "quickstart",
								},
							},
						},
					},
				},
				Port: 8080,
			},
			expectedErr: true,
		},
		{
			name: "not one service with target port",
			req: &PortForwardRequest{
				Spec: &v1.Spec{
					Resources: v1.Resources{
						{
							ID:   "v1:Service:quickstart:quickstart-dev-quickstart-private",
							Type: "Kubernetes",
							Attributes: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name":      "quickstart-dev-quickstart-private",
									"namespace": "quickstart",
								},
								"spec": map[string]interface{}{
									"ports": []interface{}{
										map[string]interface{}{
											"name":       "quickstart-dev-quickstart-private-8080-tcp",
											"port":       8888,
											"protocol":   "TCP",
											"targetPort": 8888,
										},
									},
									"selector": map[string]interface{}{
										"app.kubernetes.io/name": "quickstart",
									},
									"type": "ClusterIP",
								},
							},
						},
					},
				},
				Port: 8080,
			},
			expectedErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			bpo := &PortForwardOperation{}
			err := bpo.PortForward(tc.req)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
