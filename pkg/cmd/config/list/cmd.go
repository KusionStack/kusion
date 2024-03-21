package list

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/kubectl/pkg/util/templates"

	configutil "kusionstack.io/kusion/pkg/cmd/config/util"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/config"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`List all config items`)

		long = i18n.T(`
		This command lists all the kusion config items and their values.`)

		example = i18n.T(`
		# List config items
		kusion config list`)
	)

	cmd := &cobra.Command{
		Use:                   "list",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(Validate(args))
			util.CheckErr(Run())
			return
		},
	}
	return cmd
}

func Validate(args []string) error {
	return configutil.ValidateNoArg(args)
}

func Run() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("yaml marshal config configuration failed: %w", err)
	}
	fmt.Print(string(content))
	return nil
}
