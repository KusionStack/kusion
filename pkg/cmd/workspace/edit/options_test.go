package edit

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/workspace"
)

func TestOptions_Complete(t *testing.T) {
	testcases := []struct {
		name         string
		args         []string
		success      bool
		expectedOpts *Options
	}{
		{
			name:         "empty arg",
			args:         []string{},
			success:      false,
			expectedOpts: nil,
		},
		{
			name:         "more than one arg",
			args:         []string{"dev", "pre", "prod"},
			success:      false,
			expectedOpts: nil,
		},
		{
			name:    "successfully complete options",
			args:    []string{"dev"},
			success: true,
			expectedOpts: &Options{
				Name: "dev",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions()
			err := opts.Complete(tc.args)
			if tc.success {
				assert.True(t, reflect.DeepEqual(opts, tc.expectedOpts))
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestOptions_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		opts        *Options
		expectedErr error
	}{
		{
			name:        "empty workspace name",
			opts:        &Options{},
			expectedErr: workspace.ErrEmptyWorkspaceName,
		},
		{
			name: "valid workspace name",
			opts: &Options{
				Name: "dev",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestOptions_Run(t *testing.T) {
	opts := &Options{
		Name: "dev",
	}

	t.Run("workspace not found", func(t *testing.T) {
		mockey.PatchConvey("mock get workspace", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				Return(nil, workspace.ErrWorkspaceNotExist).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, workspace.ErrWorkspaceNotExist.Error())
		})
	})

	t.Run("failed to create temporary workspace directory", func(t *testing.T) {
		mockey.PatchConvey("mock make temp dir", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				Return(&v1.Workspace{
					Name: "dev",
				}, nil).Build()
			mockey.Mock(os.MkdirTemp).
				Return("", errors.New("failed to create temporary directory")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to create temporary directory")
		})
	})

	t.Run("failed to create the temporary workspace", func(t *testing.T) {
		mockey.PatchConvey("mock create workspace", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				Return(&v1.Workspace{
					Name: "dev",
				}, nil).Build()
			mockey.Mock((*workspace.Operator).CreateWorkspace).
				Return(errors.New("failed to create the workspace")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to create the workspace")
		})
	})

	t.Run("failed to execute the text editor", func(t *testing.T) {
		mockey.PatchConvey("mock execute text editor", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				Return(&v1.Workspace{
					Name: "dev",
				}, nil).Build()
			mockey.Mock((*exec.Cmd).Run).
				Return(errors.New("failed to run text editor")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to run text editor")
		})
	})

	t.Run("failed to validate the edited workspace config", func(t *testing.T) {
		mockey.PatchConvey("mock validate workspace", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				Return(&v1.Workspace{
					Name: "dev",
				}, nil).Build()
			mockey.Mock((*exec.Cmd).Run).Return((nil)).Build()
			mockey.Mock(workspace.ValidateWorkspace).
				Return(errors.New("failed to validate workspace")).Build()

			err := opts.Run()
			assert.ErrorContains(t, err, "failed to validate workspace")
		})
	})

	t.Run("successfully edit the workspace", func(t *testing.T) {
		mockey.PatchConvey("mock successfully edit workspace", t, func() {
			mockey.Mock(workspace.GetWorkspaceByDefaultOperator).
				Return(&v1.Workspace{
					Name: "dev",
				}, nil).Build()
			mockey.Mock((*exec.Cmd).Run).Return((nil)).Build()

			err := opts.Run()
			assert.Nil(t, err)
		})
	})
}
