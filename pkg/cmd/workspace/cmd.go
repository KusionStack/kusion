package workspace

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/workspace/create"
	"kusionstack.io/kusion/pkg/cmd/workspace/del"
	"kusionstack.io/kusion/pkg/cmd/workspace/list"
	"kusionstack.io/kusion/pkg/cmd/workspace/show"
	cmdswitch "kusionstack.io/kusion/pkg/cmd/workspace/switch"
	"kusionstack.io/kusion/pkg/cmd/workspace/update"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Workspace is a logical concept representing a target that stacks will be deployed to`)

		long = i18n.T(`
		Workspace is a logical concept representing a target that stacks will be deployed to.
		
		Workspace is managed by platform engineers, which contains a set of configurations that application developers do not want or should not concern, and is reused by multiple stacks belonging to different projects.`)
	)

	cmd := &cobra.Command{
		Use:           "workspace",
		Short:         short,
		Long:          templates.LongDesc(long),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	createCmd := create.NewCmd()
	updateCmd := update.NewCmd()
	showCmd := show.NewCmd()
	listCmd := list.NewCmd()
	delCmd := del.NewCmd()
	switchCmd := cmdswitch.NewCmd()
	cmd.AddCommand(createCmd, updateCmd, showCmd, listCmd, delCmd, switchCmd)

	return cmd
}
