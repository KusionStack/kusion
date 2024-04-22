package stack

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/stack/create"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Stack is a folder that contains a stack.yaml file within the corresponding project directory`)

		long = i18n.T(`
		Stack in Kusion is defined as any folder that contains a stack.yaml file within the corresponding project directory. 
		
		A stack provides a mechanism to isolate multiple deployments of the same application, serving with the target workspace to which an application will be deployed.`)
	)

	cmd := &cobra.Command{
		Use:           "stack",
		Short:         short,
		Long:          templates.LongDesc(long),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	createCmd := create.NewCmd()
	cmd.AddCommand(createCmd)

	return cmd
}
