package mod

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var modLong = i18n.T(`
		Commands for managing Kusion modules.

		These commands help you manage the lifecycle of Kusion modules.`)

// NewCmdMod returns an initialized Command instance for 'mod' sub command
func NewCmdMod(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "mod SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 "Manage Kusion modules",
		Long:                  modLong,
		Run:                   cmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	// add subcommands
	cmd.AddCommand(NewCmdInit(streams))
	cmd.AddCommand(NewCmdPush(streams))

	return cmd
}
