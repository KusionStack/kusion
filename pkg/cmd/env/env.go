package env

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	envShort = i18n.T(`Print Kusion environment information`)

	envLong = i18n.T(`
    Env prints Kusion environment information.

    By default env prints information as a shell script (on Windows, a batch file). If one
    or more variable names is given as arguments, env prints the value of each named variable
    on its own line.

    The --json flag prints the environment in JSON format instead of as a shell script.

    For more about environment variables, see "kusion env -h".`)

	envExample = i18n.T(`
		# Print Kusion environment information
		kusion env

		# Print Kusion environment information as JSON format
		kusion env --json`)
)

func NewCmdEnv() *cobra.Command {
	o := NewEnvOptions()

	cmd := &cobra.Command{
		Use:     "env",
		Short:   envShort,
		Long:    templates.LongDesc(envLong),
		Example: templates.Examples(envExample),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete()
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	cmd.Flags().BoolVarP(&o.envJSON, "json", "", false, i18n.T("Print the environment in JSON format"))

	return cmd
}
