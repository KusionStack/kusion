package stack

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/stack/create"
	"kusionstack.io/kusion/pkg/cmd/stack/del"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Stack is defined as any folder that contains a stack.yaml file within a project directory`)

		long = i18n.T(`
		Stack in Kusion is defined as any folder that contains a stack.yaml file within the corresponding project directory.
		
		A stack is the smallest operational unit configured and deployed independently, serving as the target workspace of an application to be deployed to.`)
	)

	cmd := &cobra.Command{
		Use:           "stack",
		Short:         short,
		Long:          templates.LongDesc(long),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	createCmd := create.NewCmd()
	delCmd := del.NewCmd()
	cmd.AddCommand(createCmd, delCmd)

	return cmd
}
