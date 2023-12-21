package util

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/workspace"
)

var (
	ErrNotOneArgs    = errors.New("only one arg accepted")
	ErrEmptyName     = errors.New("empty workspace name")
	ErrEmptyFilePath = errors.New("empty configuration file path")
)

// GetNameFromArgs returns workspace name specified by args.
func GetNameFromArgs(args []string) (string, error) {
	if len(args) != 1 {
		return "", ErrNotOneArgs
	}
	return args[0], nil
}

// ValidateName returns the workspace name is valid or not.
func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	return nil
}

// ValidateFilePath returns the configuration file path is valid or not.
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return ErrEmptyFilePath
	}
	return nil
}

// GetValidWorkspaceFromFile gets valid and structured workspace form file.
func GetValidWorkspaceFromFile(filePath, name string) (*v1.Workspace, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %w", filePath, err)
	}

	ws := &v1.Workspace{}
	if err = yaml.Unmarshal(content, ws); err != nil {
		return nil, fmt.Errorf("yaml unmarshal file %s failed: %w", filePath, err)
	}

	workspace.CompleteWorkspace(ws, name)
	if err = workspace.ValidateWorkspace(ws); err != nil {
		return nil, fmt.Errorf("invalid workspace configuration: %w", err)
	}
	return ws, nil
}
