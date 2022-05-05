package ls

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	lsShort = "List all project and stack information"

	lsLong = `
		List all project and stack information in the current directory or the
		specify directory.
		The default output is in a human friendly format, and it also supports
		a variety of formatted structure output.`

	lsExample = `
		# List all project and stack information in the current directory
		kusion ls

		# List all project and stack information in the specify directory
		kusion ls ./path/to/project_dir

		# List all project and stack information in the specify directory,
		# and output in a Tree format
		kusion ls ./path/to/project_dir --format=tree

		# List all project and stack information in the specify directory,
		# and output in a JSON format
		kusion ls ./path/to/project_dir --format=json

		# List all project and stack information in the specify directory,
		# and output in a YAML format
		kusion ls ./path/to/project_dir --format=yaml

		# List all project and stack by level, and output in a Tree format
		kusion ls ./path/to/project_dir --format=tree --level=1`
)

func NewCmdLs() *cobra.Command {
	o := NewLsOptions()

	cmd := &cobra.Command{
		Use:     "ls [WORKDIR]",
		Short:   i18n.T(lsShort),
		Long:    templates.LongDesc(i18n.T(lsLong)),
		Example: templates.Examples(i18n.T(lsExample)),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().StringVar(&o.OutputFormat, "format", "human",
		i18n.T("the output format of the project information. valid values: json, yaml, tree, human"))
	cmd.Flags().IntVarP(&o.Level, "level", "L", 2,
		i18n.T("max display depth of the project and stack tree. One of 0,1,2"))

	return cmd
}
