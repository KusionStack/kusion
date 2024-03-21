package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNameFromArgs(t *testing.T) {
	testcases := []struct {
		name        string
		args        []string
		expected    string
		expectedErr error
	}{
		{
			name:        "empty project name",
			args:        []string{},
			expected:    "",
			expectedErr: ErrEmptyName,
		},
		{
			name:        "more than one argument",
			args:        []string{"my-first-project", "my-second-project"},
			expected:    "",
			expectedErr: ErrNotOneArg,
		},
		{
			name:        "successfully get project name from args",
			args:        []string{"my-project"},
			expected:    "my-project",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GetNameFromArgs(tc.args)

			assert.Equal(t, tc.expected, actual)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	testcases := []struct {
		name        string
		projectName string
		expectedErr error
	}{
		{
			name:        "empty project name",
			projectName: "",
			expectedErr: errors.New("the project name must not be empty"),
		},
		{
			name:        "more than 100 characters",
			projectName: strings.Repeat("a", 101),
			expectedErr: errors.New("the project name must be less than 100 characters"),
		},
		{
			name:        "not match the regex",
			projectName: "my-project-^_^",
			expectedErr: errors.New("the project name can only contain alphanumeric, hyphens, underscores, and periods"),
		},
		{
			name:        "valid project name",
			projectName: "my-project",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateName(tc.projectName)

			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	testcases := []struct {
		name        string
		configPath  string
		expectedErr error
	}{
		{
			name:        "empty config path",
			configPath:  "",
			expectedErr: nil,
		},
		{
			name:        "invalid config path",
			configPath:  "/path/to/project.json",
			expectedErr: ErrNotYAMLConfig,
		},
		{
			name:        "valid config path",
			configPath:  "/path/to/project.yaml",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateConfigPath(tc.configPath)

			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestCreatProjectWithConfigFile(t *testing.T) {
	// Create temporary project directory for unit test.
	tmpProjectDir, err := os.MkdirTemp("", "kusion-test-create-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	defer os.RemoveAll(tmpProjectDir)

	tmpConfigPath := filepath.Join(tmpProjectDir, "project.yaml")
	if err = os.WriteFile(tmpConfigPath, []byte("name: my-project"), 0o644); err != nil {
		t.Fatalf("failed to create temporary config file: %v", err)
	}

	testcases := []struct {
		name        string
		projectDir  string
		configPath  string
		expectedErr error
	}{
		{
			name:        "project exists",
			projectDir:  tmpProjectDir,
			expectedErr: ErrProjectAlreadyExist,
		},
		{
			name:        "empty config path",
			projectDir:  filepath.Join(tmpProjectDir, "my-first-project"),
			expectedErr: nil,
		},
		{
			name:        "specified config path",
			projectDir:  filepath.Join(tmpProjectDir, "my-second-project"),
			configPath:  tmpConfigPath,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := CreateProjectWithConfigFile(tc.projectDir, tc.configPath)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	// Create temporary project directory for unit test.
	tmpProjectDir, err := os.MkdirTemp("", "kusion-test-delete-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	defer os.RemoveAll(tmpProjectDir)

	deletedProjectDir := filepath.Join(tmpProjectDir, "deleted")

	testcases := []struct {
		name        string
		projectDir  string
		expectedErr error
	}{
		{
			name:        "empty project directory",
			projectDir:  "",
			expectedErr: ErrEmptyName,
		},
		{
			name:        "deleted project directory",
			projectDir:  deletedProjectDir,
			expectedErr: nil,
		},
		{
			name:        "temporary project directory to delete",
			projectDir:  tmpProjectDir,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := DeleteProject(tc.projectDir)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}
