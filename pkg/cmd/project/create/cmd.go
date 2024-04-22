package create

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Create a new project`)

		long = i18n.T(`
		This command creates a new project.yaml file under the target directory which by default is the current working directory. 
		
		Note that the target directory needs to be an empty directory.`)

		example = i18n.T(`
		# Create a new project with the name of the current working directory
		mkdir my-project && cd my-project
		kusion project create
		
		# Create a new project in a specified target directory
		kusion project create --target /dir/to/projects/my-project`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "create",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: when a panic occurs, the created project.yaml file needs to be cleaned up.
			defer util.RecoverErr(&err)

			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())

			return
		},
	}

	cmd.Flags().StringVarP(&o.ProjectDir, "target", "t", "",
		i18n.T("specify the target directory"))

	return cmd
}
