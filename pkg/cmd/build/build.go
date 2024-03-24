package build

import (
	"github.com/spf13/cobra"

	"kusionstack.io/kusion/pkg/util/i18n"
)

func (o *Options) AddBuildFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.WorkDir, "workdir", "w", "",
		i18n.T("Specify the work directory"))
	cmd.Flags().StringSliceVarP(&o.Settings, "setting", "Y", []string{},
		i18n.T("Specify the command line setting files"))
	cmd.Flags().StringToStringVarP(&o.Arguments, "argument", "D", map[string]string{},
		i18n.T("Specify the top-level argument"))
	cmd.Flags().StringVarP(&o.Backend, "backend", "", "",
		i18n.T("the backend name"))
	cmd.Flags().StringVarP(&o.Backend, "workspace", "", "",
		i18n.T("the workspace name"))
}
