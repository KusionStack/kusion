package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"kusionstack.io/kusion/pkg/log"
)

const (
	ProjectYAMLFile     = "project.yaml"
	ProjectYAMLTemplate = `name: %s`
)

var (
	ErrNotEmptyDir        = errors.New("the target directory for project creation should be empty")
	ErrProjectNameEmpty   = errors.New("the project name must not be empty")
	ErrProjectNameTooLong = errors.New("the project name must be less than 100 characters")
	ErrProjectNameInvalid = errors.New("the project name can only contain alphanumeric, hyphens, underscores, and periods")
)

// Naming rules are backend-specific. However, we provide baseline sanitization for project names
// in this file. Though the backend may enforce stronger restrictions for a project name or description
// further down the line.
var (
	validProjectNameRegexp = regexp.MustCompile("^[A-Za-z0-9_.-]{1,100}$")
)

// GetDirAndName returns the rooted path and the last element of the current working directory
// for the project to be created.
func GetDirAndName() (dir, name string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get the path of the current directory: %v", err)
	}
	name = filepath.Base(dir)

	return dir, name, nil
}

// ValidateProjectDir ensures the project directory for creation is valid.
func ValidateProjectDir(dir string) error {
	_, err := os.Stat(dir)
	// If the target project directory does not exist, Kusion will do the creation.
	if os.IsNotExist(err) {
		log.Infof("Target project directory '%s' does not exist, creating it.", dir)

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read the target project directory: %v", err)
	}

	// The project directory to be created needs to be empty initially.
	if len(files) > 0 {
		return ErrNotEmptyDir
	}

	return nil
}

// ValidateProjectName ensures a project name is valid.
func ValidateProjectName(name string) error {
	if name == "" {
		return ErrProjectNameEmpty
	}

	if len(name) > 100 {
		return ErrProjectNameTooLong
	}

	if !validProjectNameRegexp.MatchString(name) {
		return ErrProjectNameInvalid
	}

	return nil
}
