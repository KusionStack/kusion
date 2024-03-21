package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const (
	suffixYAML = ".yaml"
	suffixYML  = ".yml"
	stackYAML  = "stack.yaml"
)

const (
	stackYAMLTemplate = `name: %q`
)

var (
	ErrEmptyName         = errors.New("empty stack name")
	ErrNotOneArg         = errors.New("only one argument is accepted")
	ErrNotYAMLConfig     = errors.New("only supports the stack configuration file in YAML format")
	ErrNotDirectory      = errors.New("referenced stack should be a directory")
	ErrStackAlreadyExist = errors.New("stack has already existed")
	ErrRefStackNotExist  = errors.New("referenced stack does not exist")
)

var validStackNameRegexp = regexp.MustCompile("^[A-Za-z0-9_.-]{1,100}$")

// GetNameFromArgs returns the stack name specified by args.
func GetNameFromArgs(args []string) (string, error) {
	if len(args) < 1 {
		return "", ErrEmptyName
	}

	if len(args) > 1 {
		return "", ErrNotOneArg
	}

	return args[0], nil
}

// ValidateName returns whether the stack name is valid or not.
func ValidateName(name string) error {
	if name == "" {
		return errors.New("the stack name must not be empty")
	}

	if len(name) > 100 {
		return errors.New("the stack name must be less than 100 characters")
	}

	if !validStackNameRegexp.MatchString(name) {
		return errors.New("the stack name can only contain alphanumeric, hyphens, underscores and periods")
	}

	return nil
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

// ValidateRefStackDir returns whether the referenced stack directory is valid or not.
func ValidateRefStackDir(refStackDir string) error {
	if refStackDir != "" {
		refStackInfo, err := os.Stat(refStackDir)
		if err != nil {
			return fmt.Errorf("failed to stat the reference stack directory: %w", err)
		}
		if !refStackInfo.IsDir() {
			return ErrNotDirectory
		}
	}

	return nil
}

// CreateStackWithRefAndConfigFile creates the stack with the referenced stack and the config file if specified.
func CreateStackWithRefAndConfigFile(stackDir, refStackDir, configPath string) error {
	// Check whether the target stack directory has already existed.
	if _, err := os.Stat(stackDir); !os.IsNotExist(err) {
		return ErrStackAlreadyExist
	}

	// Create the target stack directory.
	if err := os.Mkdir(stackDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create stack '%s': %w", stackDir, err)
	}

	// Check whether the referenced stack directory exists.
	if refStackDir != "" {
		if _, err := os.Stat(refStackDir); os.IsNotExist(err) {
			return ErrRefStackNotExist
		}

		// Copy files under the referenced stack to the target stack directory.
		err := filepath.Walk(refStackDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && !(info.Name() == stackYAML) {
				relPath, err := filepath.Rel(refStackDir, path)
				if err != nil {
					return err
				}

				newPath := filepath.Join(stackDir, relPath)
				if err = os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
					return err
				}

				src, err := os.Open(path)
				if err != nil {
					return err
				}
				defer src.Close()

				dst, err := os.Create(newPath)
				if err != nil {
					return err
				}
				defer dst.Close()

				if _, err = io.Copy(dst, src); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to copy files from referenced stack: %w", err)
		}
	}

	// Set the stack config file content with the default template.
	stackConfigContent := fmt.Sprintf(stackYAMLTemplate, filepath.Base(stackDir))

	// Set the stack config file content with the specified config file path.
	if configPath != "" {
		configContent, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read the specified stack config file: %w", err)
		}
		stackConfigContent = string(configContent)
	}

	// Create the 'stack.yaml' configuration file.
	stackConfigFile := filepath.Join(stackDir, stackYAML)
	if _, err := os.Stat(stackConfigFile); err == nil {
		_ = os.Remove(stackConfigFile)
	}
	if err := os.WriteFile(stackConfigFile, []byte(stackConfigContent), 0o640); err != nil {
		return fmt.Errorf("failed to create 'stack.yaml' for '%s': %w", filepath.Base(stackDir), err)
	}

	return nil
}

// DeleteStack deletes a specified stack.
func DeleteStack(stackDir string) error {
	if stackDir == "" {
		return ErrEmptyName
	}

	if _, err := os.Stat(stackDir); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(stackDir); err != nil {
		return fmt.Errorf("failed to delete stack '%s': %w", filepath.Base(stackDir), err)
	}

	return nil
}
