package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/cmd/project/util"
	"kusionstack.io/kusion/pkg/modules/generators/workload/secret"
)

func TestValidateStackName(t *testing.T) {
	testcases := []struct {
		name        string
		stackName   string
		expectedErr error
	}{
		{
			name:        "empty stack name",
			stackName:   "",
			expectedErr: ErrStackNameEmpty,
		},
		{
			name:        "more than 100 characters",
			stackName:   strings.Repeat("k", 101),
			expectedErr: ErrStackNameTooLong,
		},
		{
			name:        "invalid stack name",
			stackName:   "dev-:)",
			expectedErr: ErrStackNameInvalid,
		},
		{
			name:        "valid stack name",
			stackName:   "dev",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateStackName(tc.stackName)

			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestValidateProjectDir(t *testing.T) {
	// Create a temporary project directory for unit test.
	randomSuffix := secret.GenerateRandomString(16)
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-stack-util-"+randomSuffix)
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDirNotExists := filepath.Join(tmpTestRootDir, "my-project-not-exists")

	tmpProjectDirInvalid, err := os.MkdirTemp(tmpTestRootDir, "my-project-invalid")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}

	tmpProjectDirValid, err := os.MkdirTemp(tmpTestRootDir, "my-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}

	projectYAMLPath := filepath.Join(tmpProjectDirValid, util.ProjectYAMLFile)
	projectYAMLContent := fmt.Sprintf(util.ProjectYAMLTemplate, "my-project")
	if err = os.WriteFile(projectYAMLPath, []byte(projectYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create project.yaml file in project directory '%s': %v", tmpProjectDirValid, err)
	}

	testcases := []struct {
		name        string
		projectDir  string
		expectedErr error
	}{
		{
			name:        "target project directory not exists",
			projectDir:  tmpProjectDirNotExists,
			expectedErr: fmt.Errorf("target project directory '%s' does not exist", tmpProjectDirNotExists),
		},
		{
			name:        "target project directory invalid",
			projectDir:  tmpProjectDirInvalid,
			expectedErr: fmt.Errorf("invalid target project directory: %s", tmpProjectDirInvalid),
		},
		{
			name:        "valid target project directory",
			projectDir:  tmpProjectDirValid,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateProjectDir(tc.projectDir)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestValidateStackDir(t *testing.T) {
	// Create a temporary project and stack directory for unit tests.
	randomSuffix := secret.GenerateRandomString(16)
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-stack-util-"+randomSuffix)
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDir, err := os.MkdirTemp(tmpTestRootDir, "my-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	projectYAMLPath := filepath.Join(tmpProjectDir, util.ProjectYAMLFile)
	projectYAMLContent := fmt.Sprintf(util.ProjectYAMLTemplate, "my-project")
	if err = os.WriteFile(projectYAMLPath, []byte(projectYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create project.yaml file in project directory '%s': %v", tmpProjectDir, err)
	}

	tmpStackDirNotEmpty, err := os.MkdirTemp(tmpProjectDir, "dev")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}
	stackYAMLPath := filepath.Join(tmpStackDirNotEmpty, StackYAMLFile)
	stackYAMLContent := fmt.Sprintf(StackYAMLTemplate, "dev")
	if err = os.WriteFile(stackYAMLPath, []byte(stackYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create stack.yaml file in stack directory '%s': %v", tmpStackDirNotEmpty, err)
	}

	tmpStackDirNotExists := filepath.Join(tmpProjectDir, "prod")

	testcases := []struct {
		name        string
		stackDir    string
		expectedErr error
	}{
		{
			name:        "stack directory not empty",
			stackDir:    tmpStackDirNotEmpty,
			expectedErr: ErrNotEmptyDir,
		},
		{
			name:        "stack directory not exists",
			stackDir:    tmpStackDirNotExists,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateStackDir(tc.stackDir)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestValidateRefStackDir(t *testing.T) {
	// Create a temporary project and stack directory for unit tests.
	randomSuffix := secret.GenerateRandomString(16)
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-stack-util-"+randomSuffix)
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDir, err := os.MkdirTemp(tmpTestRootDir, "my-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	projectYAMLPath := filepath.Join(tmpProjectDir, util.ProjectYAMLFile)
	projectYAMLContent := fmt.Sprintf(util.ProjectYAMLTemplate, "my-project")
	if err = os.WriteFile(projectYAMLPath, []byte(projectYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create project.yaml file in project directory '%s': %v", tmpProjectDir, err)
	}

	tmpStackDirNotExists := filepath.Join(tmpProjectDir, "dev")

	tmpStackDirInvalid, err := os.MkdirTemp(tmpProjectDir, "pre")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}

	tmpStackDirValid, err := os.MkdirTemp(tmpProjectDir, "prod")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}
	stackYAMLPath := filepath.Join(tmpStackDirValid, StackYAMLFile)
	stackYAMLContent := fmt.Sprintf(StackYAMLTemplate, "prod")
	if err = os.WriteFile(stackYAMLPath, []byte(stackYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create stack.yaml file in stack directory '%s': %v", tmpStackDirValid, err)
	}

	testcases := []struct {
		name        string
		stackDir    string
		expectedErr error
	}{
		{
			name:        "referenced stack directory not exists",
			stackDir:    tmpStackDirNotExists,
			expectedErr: fmt.Errorf("the referenced stack directory '%s' does not exist", tmpStackDirNotExists),
		},
		{
			name:        "referenced stack directory invalid",
			stackDir:    tmpStackDirInvalid,
			expectedErr: fmt.Errorf("invalid referenced stack directory: %s", tmpStackDirInvalid),
		},
		{
			name:        "valid referenced stack directory",
			stackDir:    tmpStackDirValid,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := ValidateRefStackDir(tc.stackDir)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, actualErr, tc.expectedErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestCreateWithRefStack(t *testing.T) {
	// Create a temporary project and stack directory for unit tests.
	randomSuffix := secret.GenerateRandomString(16)
	tmpTestRootDir, err := os.MkdirTemp("", "kusion-test-stack-util-"+randomSuffix)
	if err != nil {
		t.Fatalf("failed to create temporary test root directory: %v", err)
	}
	defer os.RemoveAll(tmpTestRootDir)

	tmpProjectDir, err := os.MkdirTemp(tmpTestRootDir, "my-project")
	if err != nil {
		t.Fatalf("failed to create temporary project directory: %v", err)
	}
	projectYAMLPath := filepath.Join(tmpProjectDir, util.ProjectYAMLFile)
	projectYAMLContent := fmt.Sprintf(util.ProjectYAMLTemplate, "my-project")
	if err = os.WriteFile(projectYAMLPath, []byte(projectYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create project.yaml file in project directory '%s': %v", tmpProjectDir, err)
	}

	tmpRefStackDir, err := os.MkdirTemp(tmpProjectDir, "dev")
	if err != nil {
		t.Fatalf("failed to create temporary stack directory: %v", err)
	}
	stackYAMLPath := filepath.Join(tmpRefStackDir, StackYAMLFile)
	stackYAMLContent := fmt.Sprintf(StackYAMLTemplate, "dev")
	if err = os.WriteFile(stackYAMLPath, []byte(stackYAMLContent), 0o644); err != nil {
		t.Fatalf("failed to create stack.yaml file in stack directory '%s': %v", tmpRefStackDir, err)
	}

	kclModPath := filepath.Join(tmpRefStackDir, KCLModFile)
	kclModContent := fmt.Sprint(KCLModFileTemplate)
	if err = os.WriteFile(kclModPath, []byte(kclModContent), 0o644); err != nil {
		t.Fatalf("failed to create kcl.mod file in stack directory '%s': %v", tmpRefStackDir, err)
	}

	mainKCLPath := filepath.Join(tmpRefStackDir, MainKCLFile)
	mainKCLContent := fmt.Sprint(MainKCLFileTemplate)
	if err = os.WriteFile(mainKCLPath, []byte(mainKCLContent), 0o644); err != nil {
		t.Fatalf("failed to create main.k file in stack directory '%s': %v", tmpRefStackDir, err)
	}

	t.Run("create a new stack with referenced stack", func(t *testing.T) {
		actualErr := CreateWithRefStack("prod", filepath.Join(tmpProjectDir, "prod"), tmpRefStackDir)
		assert.NoError(t, actualErr)
	})
}
