package check

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/compile"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	checkShort = "Check if KCL configurations in current directory ok to compile"

	checkLong = `
		Check if KCL configurations in current directory ok to compile.`

	checkExample = `
		# Check configuration in main.k
		kusion check main.k

		# Check main.k with arguments
		kusion check main.k -D name=test -D age=18

		# Check main.k with arguments from settings.yaml
		kusion check main.k -Y settings.yaml

		# Check main.k with work directory
		kusion check main.k -w appops/demo/dev`
)

func NewCmdCheck() *cobra.Command {
	o := compile.NewCompileOptions()
	o.IsCheck = true

	cmd := &cobra.Command{
		Use:     "check",
		Short:   i18n.T(checkShort),
		Long:    templates.LongDesc(i18n.T(checkLong)),
		Example: templates.Examples(i18n.T(checkExample)),
		Aliases: []string{"vl"},
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddCompileFlags(cmd)

	cmd.Flags().BoolVarP(&o.DisableNone, "disable-none", "n", false,
		i18n.T("Disable dumping None values"))
	cmd.Flags().BoolVarP(&o.OverrideAST, "override-AST", "a", false,
		i18n.T("Specify the override option"))

	return cmd
}
