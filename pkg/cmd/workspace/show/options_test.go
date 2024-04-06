package show

import (
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

func TestOptions_Complete(t *testing.T) {
	testcases := []struct {
		name         string
		args         []string
		success      bool
		expectedOpts *Options
	}{
		{
			name:         "successfully complete options",
			args:         []string{"dev"},
			success:      true,
			expectedOpts: &Options{Name: "dev"},
		},
		{
			name:         "complete field invalid args",
			args:         []string{"dev", "prod"},
			success:      false,
			expectedOpts: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions()
			err := opts.Complete(tc.args)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.True(t, reflect.DeepEqual(opts, tc.expectedOpts))
			}
		})
	}
}

func TestOptions_Run(t *testing.T) {
	testcases := []struct {
		name    string
		opts    *Options
		success bool
	}{
		{
			name: "successfully run",
			opts: &Options{
				Name: "dev",
			},
			success: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock show workspace", t, func() {
				mockey.Mock(backend.NewWorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
				mockey.Mock((*workspacestorages.LocalStorage).Get).Return(&v1.Workspace{Name: "dev"}, nil).Build()

				err := tc.opts.Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
