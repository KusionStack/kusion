package resource

import (
	"encoding/json"
	"fmt"

	"github.com/liu-hm19/pterm"

	"gopkg.in/yaml.v3"

	"k8s.io/kubectl/pkg/util/templates"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kusionstack.io/kusion/pkg/backend"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	showShort = i18n.T("Show details of a resource of the current or specified stack")

	showLong = i18n.T(`
	Show details of a resource of the current or specified stack.
	
	This command displays detailed information about a resource of the current project in the current or a specified workspace
	`)

	showExample = i18n.T(`
	# Show details of a specific resource of the current project in the current workspace
	kusion resource show --id=hashicorp:viettelcloud:viettelcloud_db_instance:example-mysql

	# Show details of a specific resource with specified project and workspace
	kusion resource show --id=hashicorp:viettelcloud:viettelcloud_db_instance:example-mysql --project=example --workspace=dev
	
	# Show details of a specific resource with specified backend
	kusion resource show --id=hashicorp:viettelcloud:viettelcloud_db_instance:example-mysql --backend=local
	
	# Show details of a specific resource with specified output format
	kusion resource show --id=hashicorp:viettelcloud:viettelcloud_db_instance:example-mysql --output=json
	`)
)

// ShowFlags reflects the information that CLI is gathering via flags,
// which will be converted into ShowOptions.
type ShowFlags struct {
	ID        *string
	Project   *string
	Workspace *string
	Backend   *string
	Output    string
}

// ShowOptions defines the configuration parameters for the `kusion resource show` command.
type ShowOptions struct {
	ID             *string
	Project        *string
	Workspace      *string
	ReleaseStorage release.Storage
	Output         string
}

// NewShowFlags returns a default ShowFlags.
func NewShowFlags(_ genericiooptions.IOStreams) *ShowFlags {
	id := ""
	workspace := ""
	projectName := ""
	backendName := ""
	output := ""
	return &ShowFlags{
		ID:        &id,
		Project:   &projectName,
		Workspace: &workspace,
		Backend:   &backendName,
		Output:    output,
	}
}

// NewCmdShow creates the `kusion resource show` command.
func NewCmdShow(streams genericiooptions.IOStreams) *cobra.Command {
	flags := NewShowFlags(streams)

	cmd := &cobra.Command{
		Use:     "show",
		Short:   showShort,
		Long:    templates.LongDesc(showLong),
		Example: templates.Examples(showExample),
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

// AddFlags adds flags for a ShowOptions struct to the specified command.
func (f *ShowFlags) AddFlags(cmd *cobra.Command) {
	if f.ID != nil {
		cmd.Flags().StringVarP(f.ID, "id", "", "", i18n.T("The resource ID"))
	}
	if f.Project != nil {
		cmd.Flags().StringVarP(f.Project, "project", "", "", i18n.T("The project name"))
	}
	if f.Workspace != nil {
		cmd.Flags().StringVarP(f.Workspace, "workspace", "", "", i18n.T("The workspace name"))
	}
	if f.Backend != nil {
		cmd.Flags().StringVarP(f.Backend, "backend", "", "", i18n.T("The backend to use, supports 'local', 'oss' and 's3'"))
	}
	cmd.Flags().StringVarP(&f.Output, "output", "o", f.Output, i18n.T("Specify the output format"))
}

// ToOptions converts ShowFlags to ShowOptions.
func (f *ShowFlags) ToOptions() (*ShowOptions, error) {
	if f.ID == nil || *f.ID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}

	var storageBackend backend.Backend
	var err error
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

	workspaceStorage, err := storageBackend.WorkspaceStorage()
	if err != nil {
		return nil, err
	}
	if f.Workspace != nil && *f.Workspace != "" {
		refWorkspace, err := workspaceStorage.Get(*f.Workspace)
		if err != nil {
			return nil, err
		}
		workspaceName = refWorkspace.Name
	} else {
		currentWorkspace, err := workspaceStorage.GetCurrent()
		if err != nil {
			return nil, err
		}
		workspaceName = currentWorkspace
	}

	if f.Project != nil && *f.Project != "" {
		projectName = *f.Project
	} else {
		currentProject, _, err := project.DetectProjectAndStacks()
		if err != nil {
			return nil, err
		}
		projectName = currentProject.Name
	}
	storage, err := storageBackend.ReleaseStorage(projectName, workspaceName)
	if err != nil {
		return nil, err
	}

	return &ShowOptions{
		ID:             f.ID,
		Output:         f.Output,
		Project:        &projectName,
		Workspace:      &workspaceName,
		ReleaseStorage: storage,
	}, nil
}

// Validate checks the provided options for the `kusion resource show` command.
func (o *ShowOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `kusion resource show` command.
func (o *ShowOptions) Run() (err error) {
	rel, err := o.ReleaseStorage.Get(o.ReleaseStorage.GetLatestRevision())
	if err != nil {
		return fmt.Errorf("no release found for project: %s, workspace: %s\n",
			*o.Project, *o.Workspace)
	}

	if rel.Spec.Resources == nil || len(rel.Spec.Resources) == 0 {
		return fmt.Errorf("no resources found for project: %s, workspace: %s\n",
			*o.Project, *o.Workspace)
	}
	resourceMap := make(map[string]apiv1.Resource)
	for _, res := range rel.Spec.Resources {
		resourceMap[res.ResourceKey()] = res
	}

	stateMap := make(map[string]apiv1.Resource)
	for _, res := range rel.State.Resources {
		stateMap[res.ResourceKey()] = res
	}

	if _, ok := resourceMap[*o.ID]; !ok {
		return fmt.Errorf("no resource found for ID %s in project: %s, workspace: %s",
			*o.ID, *o.Project, *o.Workspace)
	}

	if o.Output == jsonOutput {
		resource, err := json.MarshalIndent(resourceMap[*o.ID], "", "    ")
		if err != nil {
			return err
		}
		state, err := json.MarshalIndent(stateMap[*o.ID], "", "    ")
		if err != nil {
			return err
		}
		printResourceShow(string(resource), string(state))
		return nil
	}

	resource, err := yaml.Marshal(resourceMap[*o.ID])
	if err != nil {
		return err
	}
	state, err := yaml.Marshal(stateMap[*o.ID])
	if err != nil {
		return err
	}
	printResourceShow(string(resource), string(state))

	return nil
}

func printResourceShow(resource, state string) {
	pterm.Println(pterm.Green("# Resource Spec:"))
	pterm.Println(resource)
	pterm.Println(pterm.Green("# Resource State:"))
	pterm.Println(state)
}
