package create

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Create a new stack`)

		long = i18n.T(`
		This command creates a new stack with specified name and configuration file in the YAML format or from a referenced stack.`)

		example = i18n.T(`
		# Create a new stack at current project directory
		kusion stack create my-stack
		
		# Create a new stack under the specified project directory
		kusion stack create my-stack -d /dir/to/my/projects/my-project
		
		# Create a new stack with the specified configuration file
		kusion stack create my-stack -f stack.yaml
		
		# Create a new stack from the specified referenced stack
		kusion stack create my-stack -r ref-stack`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "create",
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
	cmd.Flags().StringVarP(&o.ConfigPath, "file", "f", "", i18n.T("the path of the stack configuration file"))
	cmd.Flags().StringVarP(&o.RefStackDir, "ref", "r", "", i18n.T("the directory of the referenced stack"))

	return cmd
}
