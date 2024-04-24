package mod

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var modLong = i18n.T(`
		Commands for managing Kusion modules.

		These commands help you manage the lifecycle of Kusion modules.`)

// NewCmdMod returns an initialized Command instance for 'mod' sub command
func NewCmdMod(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "mod",
		DisableFlagsInUseLine: true,
		Short:                 "Manage Kusion modules",
		Long:                  modLong,
		Run:                   cmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	// add subcommands
	cmd.AddCommand(NewCmdInit())
	cmd.AddCommand(NewCmdPush(streams))
	cmd.AddCommand(NewCmdList(streams))

	return cmd
}
