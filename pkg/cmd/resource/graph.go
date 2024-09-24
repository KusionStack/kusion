package resource

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	graphShort = i18n.T("Display a graph of all the resources' information of the target project and target workspaces")

	graphLong = i18n.T(`
    Display information of all the resources of a project.

    This command displays information of all the resources of a project in the current or specified workspaces.
    `)

	graphExample = i18n.T(`
    # Display information of all the resources of a project in the current workspace.
    kusion resource graph --project quickstart

    # Display information of all the resources of a project in specified workspaces.
    kusion resource graph --project quickstart --workspace=dev,default
    
    # Display information of all the resource of a project in all the workspaces that has been deployed.
	kusion resource graph --project quickstart --all
	kusion resource graph --project quickstart -a
	
	# Display information of all the resource of a project with in specified workspaces with json format result.
	kusion resource graph --project quickstart --workspace dev -o json
    `)
)

const jsonOutput = "json"

var (
	// Define the width for each column to print
	idWidth              = 55
	kindWidth            = 30
	nameWidth            = 30
	cloudResourceIDWidth = 30
	statusWidth          = 30
)

// GraphFlags reflects the information that CLI is gathering via flags,
// which will be converted into GraphOptions.
type GraphFlags struct {
	Project   *string
	Workspace *[]string
	Backend   *string
	All       bool
	Output    string
}

// GraphOptions defines the configuration parameters for the `kusion release graph` command.
type GraphOptions struct {
	Project      string
	Workspace    []string
	GraphStorage map[string]graph.Storage
	Output       string
}

// NewGraphFlags returns a default GraphFlags.
func NewGraphFlags() *GraphFlags {
	projectName := ""
	workspaceName := []string{}
	backendName := ""
	all := false

	return &GraphFlags{
		Project:   &projectName,
		Workspace: &workspaceName,
		Backend:   &backendName,
		All:       all,
	}
}

// NewCmdGraph creates the `kusion resource graph` command.
func NewCmdGraph(streams genericiooptions.IOStreams) *cobra.Command {
	flags := NewGraphFlags()

	cmd := &cobra.Command{
		Use:     "graph",
		Short:   graphShort,
		Long:    templates.LongDesc(graphLong),
		Example: templates.Examples(graphExample),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o, err := flags.ToOptions()
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())

			return
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for the CLI.
func (f *GraphFlags) AddFlags(cmd *cobra.Command) {
	if f.Project != nil {
		cmd.Flags().StringVarP(f.Project, "project", "", "", i18n.T("The name of the target project"))
	}
	if f.Workspace != nil {
		cmd.Flags().StringSliceVarP(f.Workspace, "workspace", "", []string{}, i18n.T("The name of the target workspace"))
	}
	if f.Backend != nil {
		cmd.Flags().StringVarP(f.Backend, "backend", "", "", i18n.T("The backend to use, supports 'local', 'oss' and 's3'"))
	}

	cmd.Flags().StringVarP(&f.Output, "output", "o", f.Output, i18n.T("Specify the output format, json only"))
	cmd.Flags().BoolVarP(&f.All, "all", "a", false, i18n.T("Display all the resources of all the workspaces"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (f *GraphFlags) ToOptions() (*GraphOptions, error) {
	var storageBackend backend.Backend
	var err error
	// Get the backend storage
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
	projectName := ""
	workspaces := []string{}
	graphStorages := map[string]graph.Storage{}

	workspaceStorage, err := storageBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}

	if f.Project != nil && *f.Project != "" {
		projectName = *f.Project
	} else {
		return nil, fmt.Errorf("project is a must")
	}

	// Get all the available workspaces
	if f.All {
		workspaceNames, err := workspaceStorage.GetNames()
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspaceNames...)
	} else {
		// Use the workspaces that specified
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
			// If no workspace is specified, use the current workspace
		} else {
			currentWorkspace, err := workspaceStorage.GetCurrent()
			if err != nil {
				return nil, err
			}

			workspaceName = currentWorkspace
			workspaces = append(workspaces, workspaceName)
		}
	}

	// Get graph for each of the workspace
	for _, workspaceName := range workspaces {
		graphStorage, err := storageBackend.GraphStorage(projectName, workspaceName)
		if err != nil {
			return nil, err
		}
		if graphStorage.CheckGraphStorageExistence() {
			graphStorages[workspaceName] = graphStorage
		}
	}

	return &GraphOptions{
		Project:      projectName,
		Workspace:    workspaces,
		GraphStorage: graphStorages,
		Output:       f.Output,
	}, nil
}

// Validate verifies if GraphOptions are valid and without conflicts.
func (o *GraphOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `kusion resource graph` command.
func (o *GraphOptions) Run() error {
	if len(o.GraphStorage) == 0 {
		fmt.Printf("No graph found for project: %s in workspace %s", o.Project, o.Workspace)
		return nil
	}
	// Get the storage backend of the graph.
	for _, workspace := range o.Workspace {
		storage, ok := o.GraphStorage[workspace]
		if ok {
			graph, err := storage.Get()
			if err != nil {
				return err
			}

			if o.Output == jsonOutput {
				output, err := json.Marshal(graph)
				if err != nil {
					return fmt.Errorf("json marshal resource graph failed as %w", err)
				}
				fmt.Println(string(output))
			} else {
				displayGraph(graph)
			}
		}
	}

	return nil
}

// displayGraph displays resource graph
func displayGraph(graph *v1.Graph) {
	fmt.Printf("Displaying resource graph in the project %s...\n\n", graph.Project)
	fmt.Printf("Workspace: %s\n\n", graph.Workspace)

	// Print Workload Resources
	fmt.Println("Workload Resources:")
	printResourceHeader(idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth)
	for _, resource := range graph.Resources.WorkloadResources {
		printResourceRow(resource, idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth)
	}
	fmt.Println()

	// Print Dependency Resources
	fmt.Println("Dependency Resources:")
	printResourceHeader(idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth)
	for _, resource := range graph.Resources.DependencyResources {
		printResourceRow(resource, idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth)
	}
	fmt.Println()

	// Print Other Resources
	fmt.Println("Other Resources:")
	printResourceHeader(idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth)
	for _, resource := range graph.Resources.OtherResources {
		printResourceRow(resource, idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth)
	}
}

// Helper function to print the header row
func printResourceHeader(idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth int) {
	fmt.Printf("%-*s %-*s %-*s %-*s %-*s\n", idWidth, "ID", kindWidth, "Kind", nameWidth, "Name", cloudResourceIDWidth, "CloudResourceID", statusWidth, "Status")
}

// Helper function to print each row of resources with wrapping if necessary
func printResourceRow(resource *v1.GraphResource, idWidth, kindWidth, nameWidth, cloudResourceIDWidth, statusWidth int) {
	idLines := wrapText(resource.ID, idWidth)
	typeLines := wrapText(resource.Type, kindWidth)
	nameLines := wrapText(resource.Name, nameWidth)
	cloudResourceIDLines := wrapText(resource.CloudResourceID, cloudResourceIDWidth)
	statusLines := wrapText(string(resource.Status), statusWidth)

	// Find the maximum number of lines needed for this resource
	maxLines := maxNumber(len(idLines), len(typeLines), len(nameLines), len(cloudResourceIDLines), len(statusLines))

	// Print each line of the resource, line by line
	for i := 0; i < maxLines; i++ {
		id := ""
		if i < len(idLines) {
			id = idLines[i]
		}
		kind := ""
		if i < len(typeLines) {
			kind = typeLines[i]
		}
		name := ""
		if i < len(nameLines) {
			name = nameLines[i]
		}
		cloudResourceID := ""
		if i < len(cloudResourceIDLines) {
			cloudResourceID = cloudResourceIDLines[i]
		}
		status := ""
		if i < len(statusLines) {
			status = statusLines[i]
		}

		fmt.Printf("%-*s %-*s %-*s %-*s %-*s\n", idWidth, id, kindWidth, kind, nameWidth, name, cloudResourceIDWidth, cloudResourceID, statusWidth, status)
	}
}

// Helper function to wrap text based on a given width
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	for len(text) > width {
		lines = append(lines, text[:width])
		text = text[width:]
	}
	lines = append(lines, text)
	return lines
}

// Helper function to find the maximum of two integers
func maxNumber(nums ...int) int {
	if len(nums) == 0 {
		panic("max() arg is an empty sequence")
	}

	maxNum := nums[0]
	for _, num := range nums[1:] {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}
