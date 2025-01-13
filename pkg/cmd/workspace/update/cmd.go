package update

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Update a workspace configuration`)

		long = i18n.T(`
		This command updates a workspace configuration with specified configuration file, where the file must be in the YAML format.`)

		example = i18n.T(`
		# Update the current workspace
		kusion workspace update -f dev.yaml

		# Update a specified workspace and set as current
		kusion workspace update dev -f dev.yaml --current

		# Update a specified workspace in a specified backend
		kusion workspace update prod -f prod.yaml --backend oss-prod

		# Update a specified workspace with a specified name
		kusion workspace update dev --rename dev-test`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "update",
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

	cmd.Flags().StringVarP(&o.FilePath, "file", "f", "", i18n.T("the path of workspace configuration file"))
	cmd.Flags().StringVarP(&o.Backend, "backend", "b", "", i18n.T("the backend name"))
	cmd.Flags().StringVarP(&o.NewName, "rename", "r", "", i18n.T("the new name of the workspace"))
	cmd.Flags().BoolVarP(&o.Current, "current", "", false, i18n.T("set the creating workspace as current"))
	return cmd
}
