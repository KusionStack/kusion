package edit

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Edit a workspace configuration`)

		long = i18n.T(`
		This command edits a specified workspace configuraiton in text editor.`)

		example = i18n.T(`
		# Edit a workspace configuration
		kusion workspace edit dev`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "edit",
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
