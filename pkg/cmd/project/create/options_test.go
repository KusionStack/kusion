package create

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/cmd/project/util"
)

func TestOptions_Complete(t *testing.T) {
	mockedWD := "/dir/to/my/projects"

	testcases := []struct {
		name         string
		projectDir   string
		args         []string
		success      bool
		expectedOpts *Options
	}{
		{
			name:         "empty project name",
			args:         []string{},
			success:      false,
			expectedOpts: nil,
		},
		{
			name:         "more than one project name",
			args:         []string{"my-first-project", "my-second-project"},
			success:      false,
			expectedOpts: nil,
		},
		{
			name:    "current project directory",
			args:    []string{"my-project"},
			success: true,
			expectedOpts: &Options{
				Name:       "my-project",
				ProjectDir: filepath.Join(mockedWD, "my-project"),
			},
		},
		{
			name:       "specified project directory",
			projectDir: "/dir/to/specified/projects",
			args:       []string{"my-project"},
			success:    true,
			expectedOpts: &Options{
				Name:       "my-project",
				ProjectDir: filepath.Join("/dir/to/specified/projects", "my-project"),
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
			name: "invalid project name",
			opts: &Options{
				Name: "my-new-project-^_^",
			},
			success: false,
		},
		{
			name: "invalid config path",
			opts: &Options{
				Name:       "my-project",
				ConfigPath: "project.json",
			},
			success: false,
		},
		{
			name: "valid project name and config path",
			opts: &Options{
				Name:       "my-project",
				ConfigPath: "project.yaml",
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
				Name: "my-project",
			},
			success: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock create project", t, func() {
				mockey.Mock(util.CreateProjectWithConfigFile).Return(nil).Build()

				err := tc.opts.Run()
				assert.Equal(t, tc.success, err == nil)
			})
		})
	}
}
