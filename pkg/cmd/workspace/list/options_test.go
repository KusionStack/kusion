package list

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/backend"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

func TestOptions_Validate(t *testing.T) {
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
			opts := NewOptions()
			err := opts.Validate(tc.args)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestOptions_Run(t *testing.T) {
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
				mockey.Mock(backend.NewWorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
				mockey.Mock((*workspacestorages.LocalStorage).GetNames).Return([]string{"dev"}, nil).Build()
				mockey.Mock((*workspacestorages.LocalStorage).GetCurrent).Return("dev", nil).Build()

				opts := NewOptions()
				err := opts.Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
