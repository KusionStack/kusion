package del

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/cmd/stack/util"
)

func TestOptions_Complete(t *testing.T) {
	mockedWD := "/dir/to/my/projects/my-project"

	testcases := []struct {
		name         string
		projectDir   string
		args         []string
		success      bool
		expectedOpts *Options
	}{
		{
			name:         "empty stack name",
			args:         []string{},
			success:      false,
			expectedOpts: nil,
		},
		{
			name:         "more than one stack name",
			args:         []string{"dev", "pre", "prod"},
			success:      false,
			expectedOpts: nil,
		},
		{
			name:    "current project directory",
			args:    []string{"dev"},
			success: true,
			expectedOpts: &Options{
				Name:       "dev",
				ProjectDir: mockedWD,
				StackDir:   filepath.Join(mockedWD, "dev"),
			},
		},
		{
			name:       "specified project directory",
			projectDir: "/dir/to/specified/projects/my-project",
			args:       []string{"dev"},
			success:    true,
			expectedOpts: &Options{
				Name:       "dev",
				ProjectDir: "/dir/to/specified/projects/my-project",
				StackDir:   filepath.Join("/dir/to/specified/projects/my-project", "dev"),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock os.Getwd", t, func() {
				mockey.Mock(os.Getwd).To(func() (string, error) {
					return mockedWD, nil
				}).Build()

				opts := NewOptions()
				opts.ProjectDir = tc.projectDir
				err := opts.Complete(tc.args)
				assert.Equal(t, tc.success, err == nil)
				if tc.success {
					assert.True(t, reflect.DeepEqual(opts, tc.expectedOpts))
				}
			})
		})
	}
}

func TestOptions_Validate(t *testing.T) {
	testcases := []struct {
		name    string
		opts    *Options
		success bool
	}{
		{
			name: "invalid stack name",
			opts: &Options{
				Name: "dev-^_^",
			},
			success: false,
		},
		{
			name: "valid stack name",
			opts: &Options{
				Name: "dev",
			},
			success: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			assert.Equal(t, tc.success, err == nil)
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
			mockey.PatchConvey("mock delete stack", t, func() {
				mockey.Mock(util.DeleteStack).Return(nil).Build()

				err := tc.opts.Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
