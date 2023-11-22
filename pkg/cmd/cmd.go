package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/apply"
	"kusionstack.io/kusion/pkg/cmd/build"
	// we need to import the compile pkg to keep the compile command available
	"kusionstack.io/kusion/pkg/cmd/compile" //nolint:staticcheck
	"kusionstack.io/kusion/pkg/cmd/deps"
	"kusionstack.io/kusion/pkg/cmd/destroy"
	"kusionstack.io/kusion/pkg/cmd/env"
	cmdinit "kusionstack.io/kusion/pkg/cmd/init"
	"kusionstack.io/kusion/pkg/cmd/ls"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/cmd/version"
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

func NewKusionctlCmd(in io.Reader, out, err io.Writer) *cobra.Command {
	// Sending in 'nil' for the getLanguageFn() results in using LANGUAGE, LC_ALL,
	// LC_MESSAGES, or LANG environment variable in sequence.
	_ = i18n.LoadTranslations(i18n.DomainKusion, nil)

	var (
		rootShort = i18n.T(`Kusion is the platform engineering engine of KusionStack`)

		rootLong = i18n.T(`
		Kusion is the platform engineering engine of KusionStack. 
		It delivers intentions to Kubernetes, Clouds, and On-Premise resources.`)
	)

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:           "kusion",
		Short:         rootShort,
		Long:          templates.LongDesc(rootLong),
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
				build.NewCmdBuild(),
				ls.NewCmdLs(),
				deps.NewCmdDeps(),
			},
		},
		{
			Message: "RuntimeMap Commands:",
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
