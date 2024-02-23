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
		# Update a workspace configuration
		kusion workspace update dev -f dev.yaml`)
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

	return cmd
}
