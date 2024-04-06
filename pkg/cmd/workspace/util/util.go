package util

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/workspace"
	"kusionstack.io/kusion/pkg/workspace/storages"
)

var (
	ErrMoreThanOneArgs = errors.New("more than one args are not accepted")
	ErrEmptyName       = errors.New("empty workspace name")
	ErrInvalidDefault  = errors.New("invalid default workspace")
	ErrEmptyFilePath   = errors.New("empty configuration file path")
)

// GetNameFromArgs returns workspace name specified by args.
func GetNameFromArgs(args []string) (string, error) {
	if len(args) > 1 {
		return "", ErrMoreThanOneArgs
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", nil
}

// ValidateName returns the workspace name is valid or not, which is used for getting and updating workspace.
func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	return nil
}

// ValidateNotDefaultName returns true if ValidateName is true and the workspace name is not default, which is
// used for creating and deleting workspace.
func ValidateNotDefaultName(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	if name == storages.DefaultWorkspace {
		return ErrInvalidDefault
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

	ws.Name = name
	if err = workspace.ValidateWorkspace(ws); err != nil {
		return nil, fmt.Errorf("invalid workspace configuration: %w", err)
	}
	return ws, nil
}
