package list

import (
	"errors"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var ErrNotEmptyArgs = errors.New("no args accepted")

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`List all workspace names`)

		long = i18n.T(`
		This command list the names of all workspaces.`)

		example = i18n.T(`
		# List all workspace names
		kusion workspace list

		# List all workspace names in a specified backend
		kusion workspace list --backend oss-prod
`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "list",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Validate(args))
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVarP(&o.Backend, "backend", "", "", i18n.T("the backend name"))
	return cmd
}
