package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/kusionctl/cmd/apply"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/check"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/compile"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/deps"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/destroy"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/env"
	cmdinit "kusionstack.io/kusion/pkg/kusionctl/cmd/init"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/ls"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/preview"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/version"
	"kusionstack.io/kusion/pkg/util/i18n"
)

// NewDefaultKusionctlCommand creates the `kusionctl` command with default arguments
func NewDefaultKusionctlCommand() *cobra.Command {
	return NewDefaultKusionctlCommandWithArgs(os.Args, os.Stdin, os.Stdout, os.Stderr)
}

// NewDefaultKusionctlCommandWithArgs creates the `kusionctl` command with arguments
func NewDefaultKusionctlCommandWithArgs(args []string, in io.Reader, out, errOut io.Writer) *cobra.Command {
	kusionctl := NewKusionctlCmd(in, out, errOut)
	if len(args) <= 1 {
		return kusionctl
	}
	cmdPathPieces := args[1:]
	if _, _, err := kusionctl.Find(cmdPathPieces); err == nil {
		// sub command exist
		return kusionctl
	}
	return kusionctl
}

var (
	rootShort = "kusion manages the Kubernetes cluster by code"
	rootLong  = "kusion is a cloud-native programmable technology stack, which manages the Kubernetes cluster by code."
)

func NewKusionctlCmd(in io.Reader, out, err io.Writer) *cobra.Command {
	// Sending in 'nil' for the getLanguageFn() results in using
	// the LANG environment variable.
	//
	// TODO: Consider adding a flag or file preference for setting
	// the language, instead of just loading from the LANG env. variable.
	_ = i18n.LoadTranslations("kusion", nil)

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:           "kusion",
		Short:         i18n.T(rootShort),
		Long:          templates.LongDesc(i18n.T(rootLong)),
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	groups := templates.CommandGroups{
		{
			Message: "Configuration Commands:",
			Commands: []*cobra.Command{
				cmdinit.NewCmdInit(),
				compile.NewCmdCompile(),
				check.NewCmdCheck(),
				ls.NewCmdLs(),
				deps.NewCmdDeps(),
			},
		},
		{
			Message: "Runtime Commands:",
			Commands: []*cobra.Command{
				preview.NewCmdPreview(),
				apply.NewCmdApply(),
				destroy.NewCmdDestroy(),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	templates.ActsAsRootCommand(cmds, filters, groups...)
	// Add other subcommands
	// TODO: add plugin subcommand
	// cmds.AddCommand(plugin.NewCmdPlugin(f, ioStreams))
	cmds.AddCommand(version.NewCmdVersion())
	cmds.AddCommand(env.NewCmdEnv())

	return cmds
}
