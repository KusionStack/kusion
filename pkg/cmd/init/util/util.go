package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	ErrNotEmptyDir        = errors.New("not empty directory for initialization")
	ErrEmptyProjectName   = errors.New("the project name must not be empty")
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
// for the initialized demo project.
func GetDirAndName() (dir, name string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get the path of the current directory: %v", err)
	}
	name = filepath.Base(dir)

	return dir, name, nil
}

// ValidateProjectDir ensures the project directory for initialization is valid.
func ValidateProjectDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read the current directory: %v", err)
	}

	// The demo project directory to be initialized needs to be empty initially.
	if len(files) > 0 {
		return ErrNotEmptyDir
	}

	return nil
}

// ValidateProjectName ensures a project name is valid.
func ValidateProjectName(name string) error {
	if name == "" {
		return ErrEmptyProjectName
	}

	if len(name) > 100 {
		return ErrProjectNameTooLong
	}

	if !validProjectNameRegexp.MatchString(name) {
		return ErrProjectNameInvalid
	}

	return nil
}
