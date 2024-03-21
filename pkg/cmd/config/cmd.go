package config

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/config/get"
	"kusionstack.io/kusion/pkg/cmd/config/list"
	"kusionstack.io/kusion/pkg/cmd/config/set"
	"kusionstack.io/kusion/pkg/cmd/config/unset"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Interact with the Kusion config`)

		long = i18n.T(`	
		Config contains the operation of Kusion configurations.`)
	)

	cmd := &cobra.Command{
		Use:           "config",
		Short:         short,
		Long:          templates.LongDesc(long),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	getCmd := get.NewCmd()
	listCmd := list.NewCmd()
	setCmd := set.NewCmd()
	unsetCmd := unset.NewCmd()
	cmd.AddCommand(getCmd, listCmd, setCmd, unsetCmd)

	return cmd
}
