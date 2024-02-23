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
		This command deletes a specified workspace.`)

		example = i18n.T(`
		# Delete a workspace
		kusion workspace delete dev`)
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
	return cmd
}
