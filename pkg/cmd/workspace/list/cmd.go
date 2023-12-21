package list

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/workspace"
)

var ErrNotEmptyArgs = errors.New("no args accepted")

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`List all workspace names`)

		long = i18n.T(`
		This command list the names of all workspaces.`)

		example = i18n.T(`
		# List all workspace names
		kusion workspace list`)
	)

	cmd := &cobra.Command{
		Use:                   "list",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			if err != nil {
				return err
			}
			util.CheckErr(Validate(args))
			util.CheckErr(Run())
			return
		},
	}
	return cmd
}

func Validate(args []string) error {
	if len(args) != 0 {
		return ErrNotEmptyArgs
	}
	return nil
}

func Run() error {
	names, err := workspace.GetWorkspaceNamesByDefaultOperator()
	if err != nil {
		return err
	}
	content, err := yaml.Marshal(names)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace names failed: %w", err)
	}
	fmt.Print(string(content))
	return nil
}
