package mod

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type InitModOptions struct{}

var (
	initLong    = ``
	initExample = ``
)

// NewCmdInit returns an initialized Command instance for the 'mod init' sub command
func NewCmdInit(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "",
		DisableFlagsInUseLine: true,
		Short:                 "",
		Long:                  initLong,
		Example:               initExample,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	return cmd
}
