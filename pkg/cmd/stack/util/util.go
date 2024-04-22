package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"kusionstack.io/kusion/pkg/log"
)

const (
	StackYAMLFile     = "stack.yaml"
	StackYAMLTemplate = `# The metadata information of the stack. 
name: %s`

	KCLModFile         = "kcl.mod"
	KCLModFileTemplate = `# Please add the modules you need in 'dependencies'. 
[dependencies]
kam = { git = "https://github.com/KusionStack/kam.git", tag = "0.1.0" }`

	MainKCLFile         = "main.k"
	MainKCLFileTemplate = `# The configuration codes in perspective of developers.
import kam.v1.app_configuration as ac
import kam.v1.workload as wl
import kam.v1.workload.container as c

# Please replace the ${APPLICATION_NAME} with the name of your application, and complete the 
# 'AppConfiguration' instance with your own workload and accessories.
${APPLICATION_NAME}: ac.AppConfiguration {
	workload: wl.Service {
		containers: {

		}
	}
	accessories: {

	}
}`
)

var validStackNameRegexp = regexp.MustCompile("^[A-Za-z0-9_.-]{1,100}$")

var (
	ErrStackNameEmpty   = errors.New("the stack name must not be empty")
	ErrStackNameTooLong = errors.New("the stack name must be less than 100 characters")
	ErrStackNameInvalid = errors.New("the stack name can only contain alphanumeric, hyphens, underscores, and periods")
	ErrNotEmptyDir      = errors.New("not empty existing target directory for stack creation")
)

// ValidateStackName returns whether the stack name is valid or not.
func ValidateStackName(name string) error {
	if name == "" {
		return ErrStackNameEmpty
	}

	if len(name) > 100 {
		return ErrStackNameTooLong
	}

	if !validStackNameRegexp.MatchString(name) {
		return ErrStackNameInvalid
	}

	return nil
}

// ValidateProjectDir returns whether the target project directory is valid or not.
func ValidateProjectDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("target project directory '%s' does not exist", dir)
	} else {
		projectYAMLFile := filepath.Join(dir, "project.yaml")
		if _, err = os.Stat(projectYAMLFile); err != nil {
			return fmt.Errorf("invalid target project directory: %s", dir)
		}
	}

	return nil
}

// ValidateStackDir returns whether the target stack directory is valid or not.
func ValidateStackDir(dir string) error {
	_, err := os.Stat(dir)
	// If the target stack directory already exists, ensure the directory is empty.
	if !os.IsNotExist(err) {
		files, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read the existing target stack directory: %v", err)
		}

		// The target stack directory to be created needs to be empty initially.
		if len(files) > 0 {
			return ErrNotEmptyDir
		}
	} else {
		// If the target stack directory does not exist, Kusion will do the creation.
		log.Infof("Target stack directory '%s' does not exist, creating it.", dir)

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRefStackDir returns whether the referenced stack directory is valid or not.
func ValidateRefStackDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("the referenced stack directory '%s' does not exist", dir)
	} else {
		stackYAMLFile := filepath.Join(dir, StackYAMLFile)
		if _, err = os.Stat(stackYAMLFile); err != nil {
			return fmt.Errorf("invalid referenced stack directory: %s", dir)
		}
	}

	return nil
}

// CreateWithRefStack creates a new stack with referenced stack if specified.
func CreateWithRefStack(stackName, stackDir, refStackDir string) error {
	// Copy files under the referenced stack to the target stack directory.
	err := filepath.Walk(refStackDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && !(info.Name() == StackYAMLFile) {
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

	// Create the 'stack.yaml' file.
	path := filepath.Join(stackDir, StackYAMLFile)
	content := fmt.Sprintf(StackYAMLTemplate, stackName)

	return os.WriteFile(path, []byte(content), 0o644)
}
