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
	"kusionstack.io/kusion/pkg/cmd/destroy"
	"kusionstack.io/kusion/pkg/cmd/generate"
	cmdinit "kusionstack.io/kusion/pkg/cmd/init"
	"kusionstack.io/kusion/pkg/cmd/mod"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/cmd/project"
	rel "kusionstack.io/kusion/pkg/cmd/release"
	"kusionstack.io/kusion/pkg/cmd/resource"
	"kusionstack.io/kusion/pkg/cmd/server"
	"kusionstack.io/kusion/pkg/cmd/stack"
	"kusionstack.io/kusion/pkg/cmd/version"
	"kusionstack.io/kusion/pkg/cmd/workspace"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/terminal"
)

type KusionctlOptions struct {
	Arguments []string

	// UI is used to write to the CLI.
	UI *terminal.UI

	genericiooptions.IOStreams
}

// NewDefaultKusionctlCommand creates the `kusionctl` command with default arguments
func NewDefaultKusionctlCommand() *cobra.Command {
	return NewDefaultKusionctlCommandWithArgs(KusionctlOptions{
		Arguments: os.Args,
		UI:        terminal.DefaultUI(),
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
	// TODO: optimize the multi-language support.
	// Sending in 'nil' for the getLanguageFn() results in using LANGUAGE, LC_ALL,
	// LC_MESSAGES, or LANG environment variable in sequence.
	// _ = i18n.LoadTranslations(i18n.DomainKusion, nil)

	// Parent command to which all subcommands are added.
	rootCmd := &cobra.Command{
		Use: "kusion",
		Short: i18n.T(`Kusion is the Platform Orchestrator of Internal Developer Platform
		
Find more information at: https://www.kusionstack.io`),
		// 	Long: templates.LongDesc(`
		//   As a Platform Orchestrator, Kusion delivers user intentions to Kubernetes, Clouds and On-Premise infrastructures.
		//   It also enables asynchronous cooperation between the developer and the platform team, and drives separation of concerns.

		//   Find more information at:
		//   		https://www.kusionstack.io/docs/`),
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
	rootCmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	flags := rootCmd.PersistentFlags()

	addProfilingFlags(flags)

	groups := templates.CommandGroups{
		{
			Message: "Initialization Commands:",
			Commands: []*cobra.Command{
				cmdinit.NewCmd(),
			},
		},
		{
			Message: "Server Commands:",
			Commands: []*cobra.Command{
				server.NewCmdServer(),
			},
		},
		{
			Message: "Configuration Management Commands:",
			Commands: []*cobra.Command{
				config.NewCmd(),
				workspace.NewCmd(),
				project.NewCmd(),
				stack.NewCmd(),
				generate.NewCmdGenerate(o.UI, o.IOStreams),
			},
		},
		{
			Message: "Operational Commands:",
			Commands: []*cobra.Command{
				preview.NewCmdPreview(o.UI, o.IOStreams),
				apply.NewCmdApply(o.UI, o.IOStreams),
				destroy.NewCmdDestroy(o.UI, o.IOStreams),
			},
		},
		{
			Message: "Observational Commands:",
			Commands: []*cobra.Command{
				resource.NewCmdRes(o.IOStreams),
			},
		},
		{
			Message: "Module Management Commands:",
			Commands: []*cobra.Command{
				mod.NewCmdMod(o.IOStreams),
			},
		},
		{
			Message: "Release Management Commands:",
			Commands: []*cobra.Command{
				rel.NewCmdRel(o.UI, o.IOStreams),
			},
		},
	}
	groups.Add(rootCmd)

	filters := []string{"options"}

	templates.ActsAsRootCommand(rootCmd, filters, groups...)
	rootCmd.AddCommand(version.NewCmdVersion())
	rootCmd.AddCommand(options.NewCmdOptions(o.IOStreams.Out))
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd
}

func runHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}
