package del

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Delete a stack`)

		long = i18n.T(`
		This command deletes a specified stack.`)

		example = i18n.T(`
		# Delete a stack at current directory
		kusion stack delete my-stack
		
		# Delete a stack under the specified project directory
		kusion stack delete my-stack -d /dir/to/my/projects/my-project`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "delete",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())

			return
		},
	}

	cmd.Flags().StringVarP(&o.ProjectDir, "dir", "d", "", i18n.T("the parent project directory of the stack"))

	return cmd
}
