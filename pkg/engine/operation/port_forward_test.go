package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
)

func TestPortForwardOperation_PortForward(t *testing.T) {
	testcases := []struct {
		name        string
		req         *PortForwardRequest
		expectedErr error
	}{
		{
			name: "empty spec",
			req: &PortForwardRequest{
				Port: 8080,
			},
			expectedErr: ErrEmptySpec,
		},
		{
			name: "empty services",
			req: &PortForwardRequest{
				Request: models.Request{
					Intent: &v1.Spec{
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
				},
				Port: 8080,
			},
			expectedErr: ErrEmptyService,
		},
		{
			name: "not one service with target port",
			req: &PortForwardRequest{
				Request: models.Request{
					Intent: &v1.Spec{
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
				},
				Port: 8080,
			},
			expectedErr: ErrNotOneSvcWithTargetPort,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			bpo := &PortForwardOperation{}
			err := bpo.PortForward(tc.req)

			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
