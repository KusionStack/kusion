package server

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdServer() *cobra.Command {
	var (
		serverShort = i18n.T(`Start kusion server`)

		serverLong = i18n.T(`Start kusion server.`)

		serverExample = i18n.T(`
		# Start kusion server
		kusion server --mode kcp`)
	)

	o := NewServerOptions()
	cmd := &cobra.Command{
		Use:     "server",
		Short:   serverShort,
		Long:    templates.LongDesc(serverLong),
		Example: templates.Examples(serverExample),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			defer util.RecoverErr(&err)
			o.Complete(args)
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
			return
		},
	}

	o.AddServerFlags(cmd)

	return cmd
}

func (o *Options) AddServerFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Mode, "mode", "", "",
		i18n.T("Specify the mode"))
}
