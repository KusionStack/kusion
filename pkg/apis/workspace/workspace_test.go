package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockValidWorkspace() Workspace {
	return Workspace{
		Name:     "dev",
		Modules:  mockValidModuleConfigs(),
		Runtimes: mockValidRuntimeConfigs(),
		Backends: mockValidBackendConfigs(),
	}
}

func TestWorkspace_Validate(t *testing.T) {
	testcases := []struct {
		name      string
		success   bool
		workspace Workspace
	}{
		{
			name:      "valid workspace",
			success:   true,
			workspace: mockValidWorkspace(),
		},
		{
			name:      "invalid workspace empty name",
			success:   false,
			workspace: Workspace{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.workspace.Validate()
			assert.Equal(t, tc.success, err == nil)
		})
	}
}
