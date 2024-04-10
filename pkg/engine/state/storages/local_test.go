package storages

import (
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func mockState() *v1.DeprecatedState {
	return &v1.DeprecatedState{
		ID:            1,
		Project:       "wordpress",
		Stack:         "dev",
		Workspace:     "dev",
		Version:       1,
		KusionVersion: "0.11.0",
		Serial:        1,
		Operator:      "kk-confused",
		Resources: v1.Resources{
			v1.Resource{
				ID:   "v1:ServiceAccount:wordpress:wordpress",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ServiceAccount",
					"metadata": map[string]interface{}{
						"name":      "wordpress",
						"namespace": "wordpress",
					},
				},
			},
		},
	}
}

func mockStateContent() string {
	return `
id: 1
project: wordpress
stack: dev
workspace: dev
version: 1
kusionVersion: 0.11.0
serial: 1
operator: kk-confused
resources:
    - id: v1:ServiceAccount:wordpress:wordpress
      type: Kubernetes
      attributes:
        apiVersion: v1
        kind: ServiceAccount
        metadata:
            name: wordpress
            namespace: wordpress
createTime: 0001-01-01T00:00:00Z
`
}

func mockLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

func TestLocalStorage_Get(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		content []byte
		state   *v1.DeprecatedState
	}{
		{
			name:    "get local state successfully",
			success: true,
			content: []byte(mockStateContent()),
			state:   mockState(),
		},
		{
			name:    "get empty local state successfully",
			success: true,
			content: nil,
			state:   nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock read state file", t, func() {
				mockey.Mock(os.ReadFile).Return(tc.content, nil).Build()
				state, err := mockLocalStorage().Get()
				assert.Equal(t, tc.success, err == nil)
				assert.Equal(t, tc.state, state)
			})
		})
	}
}

func TestLocalStorage_Apply(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
		state   *v1.DeprecatedState
	}{
		{
			name:    "apply local state successfully",
			success: true,
			state:   mockState(),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock write state file", t, func() {
				mockey.Mock(os.WriteFile).Return(nil).Build()
				err := mockLocalStorage().Apply(tc.state)
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
