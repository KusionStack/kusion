package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/cmd/apply"
	"kusionstack.io/kusion/pkg/cmd/config"
	"kusionstack.io/kusion/pkg/cmd/generate"
	cmdinit "kusionstack.io/kusion/pkg/cmd/init"
	"kusionstack.io/kusion/pkg/cmd/mod"
	"kusionstack.io/kusion/pkg/cmd/workspace"

	"kusionstack.io/kusion/pkg/cmd/destroy"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/cmd/version"
	"kusionstack.io/kusion/pkg/util/i18n"
)

type KusionctlOptions struct {
	Arguments []string

	genericiooptions.IOStreams
}

// NewDefaultKusionctlCommand creates the `kusionctl` command with default arguments
func NewDefaultKusionctlCommand() *cobra.Command {
	return NewDefaultKusionctlCommandWithArgs(KusionctlOptions{
		Arguments: os.Args,
		IOStreams: genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	})
}

// NewDefaultKusionctlCommandWithArgs creates the `kusionctl` command with arguments
func NewDefaultKusionctlCommandWithArgs(o KusionctlOptions) *cobra.Command {
	cmd := NewKusionctlCmd(o)

	if len(o.Arguments) > 1 {
		cmdPathPieces := o.Arguments[1:]
		if _, _, err := cmd.Find(cmdPathPieces); err == nil {
			// sub command exist
			return cmd
		}
	}

	return cmd
}

func NewKusionctlCmd(o KusionctlOptions) *cobra.Command {
	// Sending in 'nil' for the getLanguageFn() results in using LANGUAGE, LC_ALL,
	// LC_MESSAGES, or LANG environment variable in sequence.
	_ = i18n.LoadTranslations(i18n.DomainKusion, nil)

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "kusion",
		Short: i18n.T(`Kusion is the Platform Orchestrator of Internal Developer Platform`),
		Long: templates.LongDesc(`
      As a Platform Orchestrator, Kusion delivers user intentions to Kubernetes, Clouds and On-Premise resources.
      Also enables asynchronous cooperation between the development and the platform team and drives separation of concerns.

      Find more information at:
            https://www.kusionstack.io/docs/user_docs/reference/cli/kusion/`),
		SilenceErrors: true,
		Run:           runHelp,
		// Hook before and after Run initialize and write profiles to disk,
		// respectively.
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			if err := flushProfiling(); err != nil {
				return err
			}
			return nil
		},
	}

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	flags := cmds.PersistentFlags()

	addProfilingFlags(flags)

	groups := templates.CommandGroups{
		{
			Message: "Configuration Commands:",
			Commands: []*cobra.Command{
				workspace.NewCmd(),
				cmdinit.NewCmd(),
				config.NewCmd(),
				generate.NewCmdGenerate(o.IOStreams),
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
		{
			Message: "Module Management Commands:",
			Commands: []*cobra.Command{
				mod.NewCmdMod(o.IOStreams),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	templates.ActsAsRootCommand(cmds, filters, groups...)
	cmds.AddCommand(version.NewCmdVersion())
	cmds.AddCommand(options.NewCmdOptions(o.IOStreams.Out))

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}
