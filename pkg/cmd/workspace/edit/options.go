package edit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"kusionstack.io/kusion/pkg/cmd/workspace/util"
	"kusionstack.io/kusion/pkg/workspace"
)

var ErrWorkspaceNameEdited = errors.New("workspace name should not be edited")

var (
	envEditor     = "EDITOR"
	defaultEditor = "vim"
	suffixYAML    = ".yaml"
)

type Options struct {
	Name string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	name, err := util.GetNameFromArgs(args)
	if err != nil {
		return err
	}
	o.Name = name
	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateName(o.Name); err != nil {
		return err
	}
	return nil
}

func (o *Options) Run() error {
	ws, err := workspace.GetWorkspaceByDefaultOperator(o.Name)
	if err != nil {
		return err
	}

	// Create the temporary workspace directory.
	tmpWorkspaceDir, err := os.MkdirTemp("", "tmp-workspace-dir")
	if err != nil {
		return fmt.Errorf("failed to create the temporary workspace directory': %v", err)
	}
	defer os.RemoveAll(tmpWorkspaceDir)

	// Create the temporary workspace file.
	tmpOperator, err := workspace.NewOperator(tmpWorkspaceDir)
	if err != nil {
		return fmt.Errorf("failed to create the temporary workspace operator: %v", err)
	}
	if err = tmpOperator.CreateWorkspace(ws); err != nil {
		return fmt.Errorf("failed to create the temporary workspace: %v", err)
	}

	// Get the text editor in the user's command-line interpreter.
	editor := os.Getenv(envEditor)
	if editor == "" {
		editor = defaultEditor
	}

	// Set the command of the text editor.
	editCmd := exec.Command(editor, filepath.Join(tmpWorkspaceDir, ws.Name+suffixYAML))
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	// Execute the text editor.
	if err = editCmd.Run(); err != nil {
		return fmt.Errorf("failed to run text editor '%s' in current command interpreter: %v", editor, err)
	}

	// Validate the edited workspace configuration.
	ws, err = tmpOperator.GetWorkspace(o.Name)
	if err != nil {
		return fmt.Errorf("failed to get the edited workspace: %v", err)
	}
	if err = workspace.ValidateWorkspace(ws); err != nil {
		return fmt.Errorf("invalid edited workspace: %v", err)
	}
	if ws.Name != o.Name {
		return ErrWorkspaceNameEdited
	}

	// Update the edited workspace with the default operator.
	if err = workspace.UpdateWorkspaceByDefaultOperator(ws); err != nil {
		return err
	}

	return nil
}
