package server

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/util/i18n"
)

func NewCmdServer() *cobra.Command {
	var (
		serverShort = i18n.T(`Start kusion server`)

		serverLong = i18n.T(`Start kusion server.`)

		serverExample = i18n.T(`
		# Start kusion server
		kusion server --mode kcp --db_host localhost:3306 --db_user root --db_pass 123456`)
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

func (o *ServerOptions) AddServerFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&o.Port, "port", "p", 80,
		i18n.T("Specify the port to listen on"))
	cmd.Flags().BoolVarP(&o.AuthEnabled, "auth-enabled", "a", false,
		i18n.T("Specify whether token authentication should be enabled"))
	cmd.Flags().StringSliceVarP(&o.AuthWhitelist, "auth-whitelist", "", []string{},
		i18n.T("Specify the list of whitelisted IAM accounts to allow access"))
	cmd.Flags().StringVarP(&o.AuthKeyType, "auth-key-type", "k", "RSA",
		i18n.T("Specify the auth key type. Default to RSA"))
	cmd.Flags().IntVarP(&o.MaxConcurrent, "max-concurrent", "", 10,
		i18n.T("Maximum number of concurrent executions including preview, apply and destroy. Default to 10."))
	cmd.Flags().StringVarP(&o.LogFilePath, "log-file-path", "", constant.DefaultLogFilePath,
		i18n.T("File path to write logs to. Default to /home/admin/logs/kusion.log"))
	o.Database.AddFlags(cmd.Flags())
	o.DefaultBackend.AddFlags(cmd.Flags())
	o.DefaultSource.AddFlags(cmd.Flags())
}
