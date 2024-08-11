package rel

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/i18n"

	cmdutil "kusionstack.io/kusion/pkg/cmd/util"

	"github.com/spf13/cobra"
)

var (
	showShort = i18n.T("Show details of a release of the current or specified stack")

	showLong = i18n.T(`
	Show details of a release of the current or specified stack.
	
	This command displays detailed information about a release of the current project in the current or a specified workspace
	`)

	showExample = i18n.T(`
	# Show details of the latest release of the current project in the current workspace
	kusion release show

	# Show details of a specific release of the current project in the current workspace
	kusion release show --revision=1

	# Show details of a specific release of the specified project in the specified workspace
	kusion release show --revision=1 --project=hoangndst --workspace=dev
	
	# Show details of the latest release with specified backend
	kusion release show --backend=local
	
	# Show details of the latest release with specified output format
	kusion release show --output=json
	`)
)

const jsonOutput = "json"

// ShowFlags reflects the information that CLI is gathering via flags,
// which will be converted into ShowOptions.
type ShowFlags struct {
	Revision  *uint64
	Project   *string
	Workspace *string
	Backend   *string
	Output    string
}

// ShowOptions defines the configuration parameters for the `kusion release show` command.
type ShowOptions struct {
	Revision       *uint64
	Project        *string
	Workspace      *string
	ReleaseStorage release.Storage
	Output         string
}

// NewShowFlags returns a default ShowFlags.
func NewShowFlags(_ genericiooptions.IOStreams) *ShowFlags {
	revision := uint64(0)
	workspace := ""
	projectName := ""
	backendName := ""
	output := ""
	return &ShowFlags{
		Revision:  &revision,
		Project:   &projectName,
		Workspace: &workspace,
		Backend:   &backendName,
		Output:    output,
	}
}

// NewCmdShow creates the `kusion release show` command.
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
	if f.Revision != nil {
		cmd.Flags().Uint64VarP(f.Revision, "revision", "", 0, i18n.T("The revision number of the release"))
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
		Revision:       f.Revision,
		Output:         f.Output,
		Project:        &projectName,
		Workspace:      &workspaceName,
		ReleaseStorage: storage,
	}, nil
}

// Validate checks the provided options for the `kusion release show` command.
func (o *ShowOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `kusion release show` command.
func (o *ShowOptions) Run() (err error) {
	var rel *v1.Release
	if o.Revision != nil && *o.Revision != 0 {
		rel, err = o.ReleaseStorage.Get(*o.Revision)
		if err != nil {
			fmt.Printf("No release found for revision %d of project: %s, workspace: %s\n",
				*o.Revision, *o.Project, *o.Workspace)
			return err
		}
	} else {
		rel, err = o.ReleaseStorage.Get(o.ReleaseStorage.GetLatestRevision())
		if err != nil {
			fmt.Printf("No release found for project: %s, workspace: %s\n",
				*o.Project, *o.Workspace)
			return err
		}
	}
	if o.Output == jsonOutput {
		data, err := json.MarshalIndent(rel, "", "    ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	} else {
		data, err := yaml.Marshal(rel)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}
