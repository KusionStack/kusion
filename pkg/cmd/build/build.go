package build

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdBuild() *cobra.Command {
	var (
		short = i18n.T(`Build Kusion modules in a Stack to the Intent`)

		long = i18n.T(`
		Build Kusion modules in a Stack to the Intent 
	
		The command must be executed in a Stack or by specifying a Stack directory with the -w flag.
		You can provide a list of arguments to replace the placeholders defined in KCL, 
		and use the --output flag to output the built results to a file`)

		example = i18n.T(`
	
		# Build main.k with arguments
		kusion build -D name=test -D age=18
	
		# Build main.k with work directory
		kusion build -w appops/demo/dev
	
		# Build configurations and write result into an output.yaml
		kusion build -o output.yaml

		# Build configurations with arguments from settings.yaml
		kusion build -Y settings.yaml
		
		# Build without output style and color
		kusion build --no-style=true`)
	)

	o := NewBuildOptions()
	cmd := &cobra.Command{
		Use:     "build",
		Short:   short,
		Long:    templates.LongDesc(long),
		Example: templates.Examples(example),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddBuildFlags(cmd)
	cmd.Flags().StringVarP(&o.Output, "output", "o", "",
		i18n.T("Specify the output file"))
	cmd.Flags().BoolVarP(&o.NoStyle, "no-style", "", false,
		i18n.T("Disable the output style and color"))

	return cmd
}

func (o *Options) AddBuildFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.WorkDir, "workdir", "w", "",
		i18n.T("Specify the work directory"))
	cmd.Flags().StringSliceVarP(&o.Settings, "setting", "Y", []string{},
		i18n.T("Specify the command line setting files"))
	cmd.Flags().StringToStringVarP(&o.Arguments, "argument", "D", map[string]string{},
		i18n.T("Specify the top-level argument"))
}
