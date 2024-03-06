package create

import (
	"errors"
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
	invalidRefStackDir := "/dir/to/ref/stack/invalid"
	validRefStackDir := "/dir/to/ref/stack/valid"

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
			name: "invalid referenced stack directory",
			opts: &Options{
				Name:        "dev",
				RefStackDir: invalidRefStackDir,
			},
			success: false,
		},
		{
			name: "invalid config path",
			opts: &Options{
				Name:       "dev",
				ConfigPath: "stack.json",
			},
			success: false,
		},
		{
			name: "valid stack name and referenced stack directory with config path",
			opts: &Options{
				Name:        "dev",
				RefStackDir: validRefStackDir,
				ConfigPath:  "stack.yaml",
			},
			success: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock validate referenced stack directory", t, func() {
				mockey.Mock(util.ValidateRefStackDir).To(func(refStackDir string) error {
					if refStackDir == validRefStackDir {
						return nil
					}

					return errors.New("invalid referenced stack directory")
				}).Build()

				err := tc.opts.Validate()
				assert.Equal(t, tc.success, err == nil)
			})
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
			mockey.PatchConvey("mock create stack", t, func() {
				mockey.Mock(util.CreateStackWithRefAndConfigFile).Return(nil).Build()

				err := tc.opts.Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
