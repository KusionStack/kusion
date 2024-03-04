package init

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Initialize the scaffolding for a demo project`)

		long = i18n.T(`
		This command initializes the scaffolding for a demo project with the name of the current directory to help users quickly get started.
		
		Note that current directory needs to be an empty directory.`)

		example = i18n.T(`
		# Initialize a demo project with the name of the current directory
		mkdir quickstart && cd quickstart
		kusion init`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:           "init",
		Short:         short,
		Long:          templates.LongDesc(long),
		Example:       templates.Examples(example),
		SilenceErrors: true,
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
