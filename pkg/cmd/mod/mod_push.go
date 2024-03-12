package mod

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type PushModOptions struct{}

var (
	pushLong    = ``
	pushExample = ``
)

// NewCmdPush returns an initialized Command instance for the 'mod push' sub command
func NewCmdPush(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "",
		DisableFlagsInUseLine: true,
		Short:                 "",
		Long:                  pushLong,
		Example:               pushExample,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	return cmd
}
