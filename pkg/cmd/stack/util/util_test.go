package util

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bytedance/mockey"
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
			name:        "empty stack name",
			args:        []string{},
			expected:    "",
			expectedErr: ErrEmptyName,
		},
		{
			name:        "more than one argument",
			args:        []string{"dev", "pre", "prod"},
			expected:    "",
			expectedErr: ErrNotOneArg,
		},
		{
			name:        "successfully get stack name from args",
			args:        []string{"dev"},
			expected:    "dev",
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
		stackName   string
		expectedErr error
	}{
		{
			name:        "empty stack name",
			stackName:   "",
			expectedErr: errors.New("the stack name must not be empty"),
		},
		{
			name:        "more than 100 characters",
			stackName:   strings.Repeat("a", 101),
			expectedErr: errors.New("the stack name must be less than 100 characters"),
		},
		{
			name:        "not match the regex",
			stackName:   "dev-^_^",
			expectedErr: errors.New("the stack name can only contain alphanumeric, hyphens, underscores and periods"),
		},
		{
			name:        "valid stack name",
			stackName:   "dev",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateName(tc.stackName)

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
			configPath:  "/path/to/stack.json",
			expectedErr: ErrNotYAMLConfig,
		},
		{
			name:        "valid config path",
			configPath:  "/path/to/stack.yaml",
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

func TestValidateRefStackDir(t *testing.T) {
	refStackDirNotExists := "/dir/to/ref/stack/not/exists"

	testcases := []struct {
		name        string
		refStackDir string
		expectedErr error
	}{
		{
			name:        "empty referenced stack directory",
			refStackDir: "",
			expectedErr: nil,
		},
		{
			name:        "referenced stack not exists",
			refStackDir: refStackDirNotExists,
			expectedErr: errors.New("failed to stat the reference stack directory"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockey.PatchConvey("mock os.Stat", t, func() {
				mockey.Mock(os.Stat).To(func(name string) (fs.FileInfo, error) {
					if name == refStackDirNotExists {
						return nil, os.ErrNotExist
					}

					return nil, nil
				}).Build()

				actualErr := ValidateRefStackDir(tc.refStackDir)
				if tc.expectedErr != nil {
					assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
				} else {
					assert.NoError(t, actualErr)
				}
			})
		})
	}
}

func TestCreateStackWithRefAndConfigFile(t *testing.T) {
	// Create temporary stack directory for unit test.
	tmpStackDir, err := os.MkdirTemp("", "kusion-test-create-stack")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}
	defer os.RemoveAll(tmpStackDir)

	tmpRefStackDir := filepath.Join(tmpStackDir, "ref")
	if err = os.Mkdir(tmpRefStackDir, 0o755); err != nil {
		t.Fatalf("failed to create temporary referenced stack directory: %v", err)
	}

	tmpConfigPath := filepath.Join(tmpRefStackDir, "stack.yaml")
	if err = os.WriteFile(tmpConfigPath, []byte("name: ref"), 0o644); err != nil {
		t.Fatalf("failed to create temporary config file: %v", err)
	}

	testcases := []struct {
		name        string
		stackDir    string
		refStackDir string
		configPath  string
		expectedErr error
	}{
		{
			name:        "existed stack directory",
			stackDir:    tmpStackDir,
			expectedErr: ErrStackAlreadyExist,
		},
		{
			name:        "empty referenced stack and config path",
			stackDir:    filepath.Join(tmpStackDir, "dev"),
			expectedErr: nil,
		},
		{
			name:        "create stack with referenced stack",
			stackDir:    filepath.Join(tmpStackDir, "pre"),
			refStackDir: tmpRefStackDir,
			expectedErr: nil,
		},
		{
			name:        "create stack with config path",
			stackDir:    filepath.Join(tmpStackDir, "gray"),
			configPath:  tmpConfigPath,
			expectedErr: nil,
		},
		{
			name:        "create stack with referenced stack and config path",
			stackDir:    filepath.Join(tmpStackDir, "prod"),
			refStackDir: tmpRefStackDir,
			configPath:  tmpConfigPath,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := CreateStackWithRefAndConfigFile(tc.stackDir, tc.refStackDir, tc.configPath)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestDeleteStack(t *testing.T) {
	// Create temporary stack directory for unit test.
	tmpStackDir, err := os.MkdirTemp("", "kusion-test-delete-stack")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}
	defer os.RemoveAll(tmpStackDir)

	deletedStackDir := filepath.Join(tmpStackDir, "deleted")

	testcases := []struct {
		name        string
		stackDir    string
		expectedErr error
	}{
		{
			name:        "empty stack directory",
			stackDir:    "",
			expectedErr: ErrEmptyName,
		},
		{
			name:        "deleted stack directory",
			stackDir:    deletedStackDir,
			expectedErr: nil,
		},
		{
			name:        "temporary stack directory to delete",
			stackDir:    tmpStackDir,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := DeleteStack(tc.stackDir)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}
