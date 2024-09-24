package project

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/project/create"
	"kusionstack.io/kusion/pkg/cmd/project/list"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Project is a folder that contains a project.yaml file and is linked to a Git repository`)

		long = i18n.T(`
		Project in Kusion is defined as any folder that contains a project.yaml file and is linked to a Git repository. 
		
		Project organizes logical configurations for internal components to orchestrate the application and assembles them to suit different roles, such as developers and platform engineers.`)
	)

	cmd := &cobra.Command{
		Use:           "project",
		Short:         short,
		Long:          templates.LongDesc(long),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	createCmd := create.NewCmd()
	listCmd := list.NewCmd()
	cmd.AddCommand(createCmd, listCmd)

	return cmd
}
