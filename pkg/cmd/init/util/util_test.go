package util

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestGetDirAndName(t *testing.T) {
	t.Run("failed to get the path of the current directory", func(t *testing.T) {
		mockey.PatchConvey("mock os.Getwd", t, func() {
			mockey.Mock(os.Getwd).To(func() (dir string, err error) {
				return "", errors.New("failed to get the path of the current directory")
			}).Build()

			_, _, err := GetDirAndName()
			assert.ErrorContains(t, err, "failed to get the path of the current directory")
		})
	})

	t.Run("successfully get directory and name", func(t *testing.T) {
		mockey.PatchConvey("mock os.Getwd", t, func() {
			mockey.Mock(os.Getwd).To(func() (dir string, err error) {
				return "/dir/to/demo-project", nil
			}).Build()

			dir, name, err := GetDirAndName()
			assert.Nil(t, err)
			assert.Equal(t, "/dir/to/demo-project", dir)
			assert.Equal(t, "demo-project", name)
		})
	})
}

func TestValidateProjectDir(t *testing.T) {
	// Create temporary project directory for unit test.
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-init-util")
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDir, err := os.MkdirTemp(tmpTestRootDir, "quickstart-test")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}

	testcases := []struct {
		name        string
		dir         string
		expectedErr error
	}{
		{
			name:        "directory not exists",
			dir:         "/dir/to/project/not/exists",
			expectedErr: errors.New("failed to read the current directory"),
		},
		{
			name:        "directory not empty",
			dir:         tmpTestRootDir,
			expectedErr: ErrNotEmptyDir,
		},
		{
			name:        "successfully validate the project directory",
			dir:         tmpProjectDir,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateProjectDir(tc.dir)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	testcases := []struct {
		name        string
		projectName string
		expectedErr error
	}{
		{
			name:        "empty project name",
			projectName: "",
			expectedErr: ErrEmptyProjectName,
		},
		{
			name:        "more than 100 characters",
			projectName: strings.Repeat("a", 101),
			expectedErr: ErrProjectNameTooLong,
		},
		{
			name:        "not match the regex",
			projectName: "quickstart-^_^",
			expectedErr: ErrProjectNameInvalid,
		},
		{
			name:        "valid project name",
			projectName: "quickstart",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateProjectName(tc.projectName)

			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}
