package resource

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var resLong = i18n.T(`
		Commands for observing Kusion resources.
		
		These commands help you observe the information of Kusion resources within a project. `)

// NewCmdRes returns an initialized Command instance for 'resource' sub command.
func NewCmdRes(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "resource",
		DisableFlagsInUseLine: true,
		Short:                 "Observe Kusion resource information",
		Long:                  templates.LongDesc(resLong),
		Run:                   cmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	cmd.AddCommand(NewCmdGraph(streams))

	return cmd
}
