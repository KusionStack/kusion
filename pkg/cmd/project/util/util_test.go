package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/modules/generators/workload/secret"
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
				return "/dir/to/my-project", nil
			}).Build()

			dir, name, err := GetDirAndName()
			assert.Nil(t, err)
			assert.Equal(t, "/dir/to/my-project", dir)
			assert.Equal(t, "my-project", name)
		})
	})
}

func TestValidateProjectDir(t *testing.T) {
	// Create a temporary project directory for unit test.
	randomSuffix := secret.GenerateRandomString(16)
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-project-util-"+randomSuffix)
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDir, err := os.MkdirTemp(tmpTestRootDir, "my-project")
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
			dir:         filepath.Join("kusion-test", secret.GenerateRandomString(32)),
			expectedErr: nil,
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
			expectedErr: ErrProjectNameEmpty,
		},
		{
			name:        "more than 100 characters",
			projectName: strings.Repeat("k", 101),
			expectedErr: ErrProjectNameTooLong,
		},
		{
			name:        "invalid project name",
			projectName: "my-project-:)",
			expectedErr: ErrProjectNameInvalid,
		},
		{
			name:        "valid project name",
			projectName: "my-project",
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
