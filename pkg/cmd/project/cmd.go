package project

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"kusionstack.io/kusion/pkg/cmd/project/create"
	"kusionstack.io/kusion/pkg/cmd/project/del"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Project is defined as any folder that contains a project.yaml file and typically linked to a git repo`)

		long = i18n.T(`
		Project in Kusion is defined as any folder that contains a project.yaml file and is typically linked to a git repository.
		
		A Project consists of one or more applications, whose purpose is to bundle application configurations and refer to a git repository.`)
	)

	cmd := &cobra.Command{
		Use:           "project",
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
