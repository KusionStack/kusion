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
		This command gets a specified workspace configuration.`)

		example = i18n.T(`
		# Show a workspace configuration
		kusion workspace show dev`)
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
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}
	return cmd
}
