package create

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Create a new workspace`)

		long = i18n.T(`
		This command creates a workspace with specified name and configuration file, where the file must be in the YAML format.`)

		example = i18n.T(`
		# Create a new workspace
		kusion workspace create dev -f dev.yaml`)
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
			if err != nil {
				return err
			}
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(&o.FilePath, "file", "f", "", i18n.T("the path of workspace configuration file"))
	return cmd
}
