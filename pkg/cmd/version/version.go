package version

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var (
	versionShort = "Print the kusion version info"

	versionLong = "Print the kusion version information for the current context."

	versionExample = `
		# Print the kusion version
		kusion version`
)

func NewCmdVersion() *cobra.Command {
	o := NewVersionOptions()

	cmd := &cobra.Command{
		Use:     "version",
		Short:   i18n.T(versionShort),
		Long:    templates.LongDesc(i18n.T(versionLong)),
		Example: templates.Examples(i18n.T(versionExample)),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete()
			util.CheckErr(o.Validate())
			o.Run()
			return
		},
	}

	cmd.Flags().BoolVarP(&o.ExportJSON, "json", "j", false,
		i18n.T("print version info as JSON"))
	cmd.Flags().BoolVarP(&o.ExportYAML, "yaml", "y", false,
		i18n.T("print version info as YAML"))
	cmd.Flags().BoolVarP(&o.Short, "short", "s", false,
		i18n.T("print version info as versionShort string"))

	return cmd
}
