package list

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"kusionstack.io/kusion/pkg/backend"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var ErrNotEmptyArgs = errors.New("no args accepted")

// Options defines the configurations for the `list` command.
// Options reflects the runtime requirements for the `list` command.
type Options struct {
	projects         map[string][]string
	Workspace        []string
	CurrentWorkspace string
}

// Flags defines the flags for the `list` command.
// Flags reflects the information gathered by the `list` command.
type Flags struct {
	Backend   *string
	Workspace *[]string
	All       bool
}

// NewFlags returns a new Flags with default values.
func NewFlags() *Flags {
	backend := ""
	workspace := []string{}
	all := false

	return &Flags{
		Backend:   &backend,
		Workspace: &workspace,
		All:       all,
	}
}

// AddFlags registers flags for the `list` command.
func (f *Flags) AddFlags(cmd *cobra.Command) {
	if f.Workspace != nil {
		cmd.Flags().StringSliceVarP(f.Workspace, "workspace", "", []string{}, i18n.T("The name of the target workspace"))
	}
	if f.Backend != nil {
		cmd.Flags().StringVarP(f.Backend, "backend", "", "", i18n.T("The backend to use, supports 'local', 'oss' and 's3'"))
	}

	cmd.Flags().BoolVarP(&f.All, "all", "a", false, i18n.T("List all the projects in all the workspaces"))
}

// ToOptions converts the Flags to the Options.
func (f *Flags) ToOptions() (*Options, error) {
	var storageBackend backend.Backend
	var err error
	// Get backend storage.
	if f.Backend != nil && *f.Backend != "" {
		storageBackend, err = backend.NewBackend(*f.Backend)
		if err != nil {
			return nil, err
		}
	} else {
		storageBackend, err = backend.NewBackend("")
		if err != nil {
			return nil, err
		}
	}

	workspaceName := ""
	workspaces := []string{}
	projects, err := storageBackend.ProjectStorage()
	if err != nil {
		return nil, err
	}
	workspaceStorage, err := storageBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}
	currentWorkspaceName, err := workspaceStorage.GetCurrent()
	if err != nil {
		return nil, err
	}

	// Get all the available workspaces if needed.
	if f.All {
		workspaceNames, err := workspaceStorage.GetNames()
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspaceNames...)
	} else {
		// Get the specified workspaces.
		if len(*f.Workspace) != 0 {
			for _, workspace := range *f.Workspace {
				if workspace != "" {
					refWorkspace, err := workspaceStorage.Get(workspace)
					if err != nil {
						return nil, err
					}
					workspaceName = refWorkspace.Name
					workspaces = append(workspaces, workspaceName)
				}
			}
			// No workspace specified, use the current workspace.
		} else {
			workspaceName = currentWorkspaceName
			workspaces = append(workspaces, workspaceName)
		}
	}

	return &Options{
		projects:         projects,
		Workspace:        workspaces,
		CurrentWorkspace: currentWorkspaceName,
	}, err
}

// Validate checks the options to see if they are valid.
func (o *Options) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `list` command.
func (o *Options) Run() error {
	if o.Workspace == nil || len(o.Workspace) == 0 {
		return fmt.Errorf("workspace is empty")
	}

	if len(o.projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	// Function to print each workspace and its projects.
	printProjects := func(workspace string, isCurrent bool) {
		if isCurrent {
			fmt.Println(pretty.GreenBold("Current Workspace: %s", workspace))
		} else {
			fmt.Printf("Workspace: %s\n", workspace)
		}

		projects, exists := o.projects[workspace]
		if !exists || len(projects) == 0 {
			fmt.Println("  No projects found")
		} else {
			// Print each project
			for _, project := range projects {
				fmt.Printf("  - %s\n", project)
			}
		}

		fmt.Println(strings.Repeat("-", 40))
	}

	// Check if CurrentWorkspace exists in the list.
	currentWorkspaceExists := false
	for _, workspace := range o.Workspace {
		if workspace == o.CurrentWorkspace {
			currentWorkspaceExists = true
			break
		}
	}

	// If CurrentWorkspace exists, print it first.
	if currentWorkspaceExists {
		printProjects(o.CurrentWorkspace, true)
	}

	// Print the other workspaces.
	for _, workspace := range o.Workspace {
		if workspace != o.CurrentWorkspace {
			printProjects(workspace, false)
		}
	}

	return nil
}
