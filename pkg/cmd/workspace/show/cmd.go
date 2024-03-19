package show

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Show a workspace configuration`)

		long = i18n.T(`
		This command gets the current or a specified workspace configuration.`)

		example = i18n.T(`
		# Show current workspace configuration
		kusion workspace show

		# Show a specified workspace configuration
		kusion workspace show dev

		# Show a specified workspace in a specified backend
		kusion workspace show prod --backend oss-prod`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "show",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(&o.Backend, "backend", "", "", i18n.T("the backend name"))
	return cmd
}
