package generate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"kusionstack.io/kusion/pkg/cmd/generate/generator"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
)

var (
	generateLong = ``

	generateExample = ``
)

// GenerateFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// GenerateOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type GenerateFlags struct {
	WorkDir string
	Values  []string
	Output  string

	genericiooptions.IOStreams
}

// GenerateOptions defines flags and other configuration parameters for the `generate` command.
type GenerateOptions struct {
	WorkDir string
	Output  string

	Generator generator.Generator
}

// NewGenerateFlags returns a default GenerateFlags
func NewGenerateFlags(streams genericiooptions.IOStreams) *GenerateFlags {
	return &GenerateFlags{
		IOStreams: streams,
	}
}

// NewCmdGenerate creates the `generate` command.
func NewCmdGenerate(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewGenerateFlags(ioStreams)

	cmd := &cobra.Command{
		Use:     "generate (-w DIRECTORY)",
		Short:   "Generate versioned Spec of target Stack",
		Long:    generateLong,
		Example: generateExample,
		Run: func(cmd *cobra.Command, args []string) {
			o, err := flags.ToOptions(cmd, args)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for a cli.
func (flags *GenerateFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flags.WorkDir, "workdir", "w", "", "The working directory for generate (default is current dir where executed)")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "File to write generated Kusion spec to")
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *GenerateFlags) ToOptions(cmd *cobra.Command, args []string) (*GenerateOptions, error) {
	// If working directory not specified, use current dir where executed
	workDir := flags.WorkDir
	if len(workDir) == 0 {
		workDir, _ = os.Getwd()
	}

	o := &GenerateOptions{
		WorkDir: workDir,
		Output:  flags.Output,
	}

	return o, nil
}

// Validate verifies if GenerateOptions are valid and without conflicts.
func (o *GenerateOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `generate` command.
func (o *GenerateOptions) Run() error {
	intent, err := o.Generator.Generate(o.WorkDir, nil)
	if err != nil {
		return err
	}
	fmt.Println(intent.Resources)
	return nil
}
