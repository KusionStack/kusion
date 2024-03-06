package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"kusionstack.io/kusion/pkg/scaffold"
)

const (
	suffixYAML  = ".yaml"
	suffixYML   = ".yml"
	projectYAML = "project.yaml"
)

const (
	projectYAMLTemplate = `name: %q`
)

var (
	ErrEmptyName           = errors.New("empty project name")
	ErrNotOneArg           = errors.New("only one argument is accepted")
	ErrNotYAMLConfig       = errors.New("only supports the project configuration file in YAML format")
	ErrProjectAlreadyExist = errors.New("project has already existed")
)

// GetNameFromArgs returns the project name specified by args.
func GetNameFromArgs(args []string) (string, error) {
	if len(args) < 1 {
		return "", ErrEmptyName
	}

	if len(args) > 1 {
		return "", ErrNotOneArg
	}

	return args[0], nil
}

// ValidateName returns whether the project name is valid or not.
func ValidateName(name string) error {
	return scaffold.ValidateProjectName(name)
}

// ValidateConfigPath returns whether the configuration file path is valid or not.
func ValidateConfigPath(configPath string) error {
	if configPath != "" {
		ext := filepath.Ext(configPath)
		if ext != suffixYAML && ext != suffixYML {
			return ErrNotYAMLConfig
		}
	}

	return nil
}

// CreateProjectWithConfigFile creates the project with the config file if specified.
func CreateProjectWithConfigFile(projectDir, configPath string) error {
	// Check whether the target project directory has already existed.
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return ErrProjectAlreadyExist
	}

	// Create the target project directory.
	if err := os.Mkdir(projectDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create project '%s': %w", projectDir, err)
	}

	// Set the project config file content with the default template.
	projectConfigContent := fmt.Sprintf(projectYAMLTemplate, filepath.Base(projectDir))

	// Set the project config file content with the specified config file path.
	if configPath != "" {
		configContent, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read the specified project config file: %w", err)
		}
		projectConfigContent = string(configContent)
	}

	// Create the 'project.yaml' configuration file.
	projectConfigFile := filepath.Join(projectDir, projectYAML)
	if err := os.WriteFile(projectConfigFile, []byte(projectConfigContent), 0o640); err != nil {
		return fmt.Errorf("failed to create 'project.yaml' for '%s': %w", filepath.Base(projectDir), err)
	}

	return nil
}

// DeleteProject deletes a specified project.
func DeleteProject(projectDir string) error {
	if projectDir == "" {
		return ErrEmptyName
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("failed to delete project '%s': %w", filepath.Base(projectDir), err)
	}

	return nil
}
