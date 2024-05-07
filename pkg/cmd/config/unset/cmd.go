package unset

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmd() *cobra.Command {
	var (
		short = i18n.T(`Unset a config item`)

		long = i18n.T(`
		This command unsets a specified kusion config item, where the config item must be registered.`)

		example = i18n.T(`
		# Unset a config item
		kusion config unset backends.s3-pre.configs.bucket`)
	)

	o := NewOptions()
	cmd := &cobra.Command{
		Use:                   "unset",
		Short:                 short,
		Long:                  templates.LongDesc(long),
		Example:               templates.Examples(example),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			util.CheckErr(o.Complete(args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}
	return cmd
}
