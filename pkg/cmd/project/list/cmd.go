package list

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

// NewCmd creates the `list` command.
func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`List the applied projects`)

		long = i18n.T(`
		This command lists all the applied projects in the target backend and target workspace. 
		
		By default list the projects in the current backend and current workspace.`)

		example = i18n.T(`
		# List the applied project in the current backend and current workspace
		kusion project list
		
		# List the applied project in a specified backend and current workspace
		kusion project list --backend default
		
		# List the applied project in a specified backend and specified workspaces
		kusion project list --backend default --workspace dev,default
		# List the applied project in a specified backend and all the workspaces
		kusion project list --backend default --all`)
	)

	flags := NewFlags()
	cmd := &cobra.Command{
		Use:                   "list",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
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
