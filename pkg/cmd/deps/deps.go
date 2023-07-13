package deps

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	depsShort = i18n.T(`Show KCL file dependency information`)

	depsLong = i18n.T(`
		Show the KCL file dependency information in the current directory or the specified workdir.
    By default, it will list all the KCL files that are dependent on the given package path.`)

	depsExample = i18n.T(`
		# List all the KCL files that are dependent by the given focus paths
        kusion deps --focus path/to/focus1 --focus path/to/focus2

		# List all the projects that depend on the given focus paths
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2

		# List all the stacks that depend on the given focus paths
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2 --only stack

		# List all the projects that depend on the given focus paths, ignoring some paths from entrance files in each stack
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2 --ignore path/to/ignore`)
)

func NewCmdDeps() *cobra.Command {
	o := NewDepsOptions()

	cmd := &cobra.Command{
		Use:     "deps [WORKDIR]",
		Short:   depsShort,
		Long:    templates.LongDesc(depsLong),
		Example: templates.Examples(depsExample),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVar(&o.Direct, "direct", "up",
		i18n.T("Inspect direct of the dependency information. Valid values: up, down. Defaults to up"))
	cmd.Flags().StringSliceVar(&o.Focus, "focus", nil,
		i18n.T("Paths to focus on to inspect. It cannot be empty and each path needs to be a valid relative path from the workdir"))
	cmd.Flags().StringVar(&o.Only, "only", "project",
		i18n.T("Direct is set to \"down\", \"only\" means only the downstream project/stack list will be output. Valid values: project, stack. Defaults to project"))
	cmd.Flags().StringSliceVar(&o.Ignore, "ignore", nil,
		i18n.T("File paths to ignore when filtering the affected stacks/projects. Each path needs to be a valid relative path from the workdir. If not set, no paths will be ignored."))

	return cmd
}
