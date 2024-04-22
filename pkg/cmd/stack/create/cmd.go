package create

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Create a new stack`)

		long = i18n.T(`
		This command creates a new stack under the target directory which by default is the current working directory. 
		
		The stack folder to be created contains 'stack.yaml', 'kcl.mod' and 'main.k' with the specified values. 
		
		Note that the target directory needs to be a valid project directory with project.yaml file`)

		example = i18n.T(`
		# Create a new stack at current project directory
		kusion stack create dev
		
		# Create a new stack in a specified target project directory
		kusion stack create dev --target /dir/to/projects/my-project
		
		# Create a new stack copied from the referenced stack under the target project directory
		kusion stack create prod --copy-from dev`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "create",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: when a panic occurs, the created stack folder needs to be cleaned up.
			defer util.RecoverErr(&err)

			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())

			return
		},
	}

	cmd.Flags().StringVarP(&o.ProjectDir, "target", "t", "",
		i18n.T("specify the target project directory"))
	cmd.Flags().StringVarP(&o.CopyFrom, "copy-from", "", "",
		i18n.T("specify the referenced stack path to copy from"))

	return cmd
}
