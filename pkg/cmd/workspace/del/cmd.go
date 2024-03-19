package del

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Delete a workspace`)

		long = i18n.T(`
		This command deletes the current or a specified workspace.`)

		example = i18n.T(`
		# Delete the current workspace
		kusion workspace delete

		# Delete a specified workspace
		kusion workspace delete dev

		# Delete a specified workspace in a specified backend
		kusion workspace delete prod --backend oss-prod`)
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
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(&o.Backend, "backend", "", "", i18n.T("the backend name"))
	return cmd
}
