package cmdswitch

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Switch the current workspace`)

		long = i18n.T(`
		This command switches the workspace, where the workspace must be created.`)

		example = i18n.T(`
		# Switch the current workspace
		kusion workspace switch dev

		# Switch the current workspace in a specified backend
		kusion workspace switch prod --backend oss-prod
`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "switch",
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

	cmd.Flags().StringVarP(&o.Backend, "backend", "", "", i18n.T("the backend name"))
	return cmd
}
