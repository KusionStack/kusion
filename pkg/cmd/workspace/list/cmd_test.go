package list

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/workspace"
)

func TestNewCmd(t *testing.T) {
	t.Run("successfully list workspace", func(t *testing.T) {
		mockey.PatchConvey("mock cmd", t, func() {
			mockey.Mock(Validate).Return(nil).Build()
			mockey.Mock(Run).Return(nil).Build()

			cmd := NewCmd()
			err := cmd.Execute()
			assert.Nil(t, err)
		})
	})
}

func TestValidate(t *testing.T) {
	testcases := []struct {
		name    string
		args    []string
		success bool
	}{
		{
			name:    "valid args",
			args:    nil,
			success: true,
		},
		{
			name:    "invalid args",
			args:    []string{"dev"},
			success: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.args)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestRun(t *testing.T) {
	testcases := []struct {
		name    string
		success bool
	}{
		{
			name:    "successfully run",
			success: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock get workspace names", t, func() {
				mockey.Mock(workspace.GetWorkspaceNamesByDefaultOperator).
					Return([]string{"dev"}, nil).
					Build()

				err := Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
