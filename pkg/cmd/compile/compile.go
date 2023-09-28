package compile

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdCompile() *cobra.Command {
	var (
		compileShort = i18n.T(`Compile Kusion models to the Spec (intent)`)

		compileLong = i18n.T(`
		Compile Kusion models in a Stack to the Spec (intent)
	
		The command must be executed in a Stack or specifying a Stack dir with the -w flag. 
		You can specify a list of arguments to replace the placeholders defined in KCL,
		and output the compiled results to a file when using --output flag.`)

		compileExample = i18n.T(`
	
		# Compile main.k with arguments
		kusion compile -D name=test -D age=18
	
		# Compile main.k with arguments from settings.yaml
		kusion compile -Y settings.yaml
	
		# Compile main.k with work directory
		kusion compile -w appops/demo/dev
	
		# Compile with override
		kusion compile -O __main__:appConfiguration.image=nginx:latest -a
	
		# Compile main.k and write result into output.yaml
		kusion compile -o output.yaml
		
		# Compile without output style and color
		kusion compile --no-style=true`)
	)

	o := NewCompileOptions()
	cmd := &cobra.Command{
		Use:     "compile",
		Short:   compileShort,
		Long:    templates.LongDesc(compileLong),
		Example: templates.Examples(compileExample),
		Aliases: []string{"cl"},
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddCompileFlags(cmd)
	cmd.Flags().StringVarP(&o.Output, "output", "o", "",
		i18n.T("Specify the output file"))
	cmd.Flags().BoolVarP(&o.DisableNone, "disable-none", "n", false,
		i18n.T("Disable dumping None values"))
	cmd.Flags().BoolVarP(&o.OverrideAST, "override-AST", "a", false,
		i18n.T("Specify the override option"))
	cmd.Flags().BoolVarP(&o.NoStyle, "no-style", "", false,
		i18n.T("Disable the output style and color"))

	return cmd
}

func (o *Options) AddCompileFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.WorkDir, "workdir", "w", "",
		i18n.T("Specify the work directory"))
	cmd.Flags().StringSliceVarP(&o.Settings, "setting", "Y", []string{},
		i18n.T("Specify the command line setting files"))
	cmd.Flags().StringToStringVarP(&o.Arguments, "argument", "D", map[string]string{},
		i18n.T("Specify the top-level argument"))
	cmd.Flags().StringSliceVarP(&o.Overrides, "overrides", "O", []string{},
		i18n.T("Specify the configuration override path and value"))
}
