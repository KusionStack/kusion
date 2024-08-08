package rel

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/meta"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	listShort = i18n.T("List all releases of the current stack")

	listLong = i18n.T(`
    List all releases of the current stack.

    This command displays information about all releases of the current stack in the current or a specified workspace,
    including their revision, phase, and creation time.
    `)

	listExample = i18n.T(`
    # List all releases of the current stack in current workspace
    kusion release list

    # List all releases of the current stack in a specified workspace
    kusion release list --workspace=dev
    `)
)

// ListFlags reflects the information that CLI is gathering via flags,
// which will be converted into ListOptions.
type ListFlags struct {
	MetaFlags *meta.MetaFlags
}

// ListOptions defines the configuration parameters for the `kusion release list` command.
type ListOptions struct {
	*meta.MetaOptions
}

// NewListFlags returns a default ListFlags.
func NewListFlags(streams genericiooptions.IOStreams) *ListFlags {
	return &ListFlags{
		MetaFlags: meta.NewMetaFlags(),
	}
}

// NewCmdList creates the `kusion release list` command.
func NewCmdList(streams genericiooptions.IOStreams) *cobra.Command {
	flags := NewListFlags(streams)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   listShort,
		Long:    templates.LongDesc(listLong),
		Example: templates.Examples(listExample),
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
func (f *ListFlags) AddFlags(cmd *cobra.Command) {
	f.MetaFlags.AddFlags(cmd)
}

// ToOptions converts from CLI inputs to runtime inputs.
func (f *ListFlags) ToOptions() (*ListOptions, error) {
	metaOpts, err := f.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &ListOptions{
		MetaOptions: metaOpts,
	}

	return o, nil
}

// Validate verifies if ListOptions are valid and without conflicts.
func (o *ListOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `kusion release list` command.
func (o *ListOptions) Run() error {
	// Get the storage backend of the release.
	storage, err := o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return err
	}

	// Get all releases.
	releases := storage.GetRevisions()
	if len(releases) == 0 {
		fmt.Printf("No releases found for project: %s, workspace: %s\n",
			o.RefProject.Name, o.RefWorkspace.Name)
		return nil
	}

	// Print the releases
	fmt.Printf("Releases for project: %s, workspace: %s\n\n", o.RefProject.Name, o.RefWorkspace.Name)
	fmt.Printf("%-10s %-15s %-30s\n", "Revision", "Phase", "Creation Time")
	fmt.Println("------------------------------------------------------")
	for _, revision := range releases {
		r, err := storage.Get(revision)
		if err != nil {
			return err
		}
		fmt.Printf("%-10d %-15s %-30s\n", r.Revision, string(r.Phase), r.CreateTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}
