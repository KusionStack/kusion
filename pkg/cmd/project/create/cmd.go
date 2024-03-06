package create

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Create a new project`)

		long = i18n.T(`
		This command creates a new project with specified name and configuration file in the YAML format.`)

		example = i18n.T(`
		# Create a new project at current directory
		kusion project create my-project
		
		# Create a new project under the specified directory
		kusion project create my-project -d /dir/to/my/projects
		
		# Create a new project with the specified configuration file
		kusion project create my-project -f project.yaml`)
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

	cmd.Flags().StringVarP(&o.ProjectDir, "dir", "d", "", i18n.T("the parent directory of the project"))
	cmd.Flags().StringVarP(&o.ConfigPath, "file", "f", "", i18n.T("the path of project configuration file"))

	return cmd
}
