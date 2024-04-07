package generate

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/generate/generator"
	"kusionstack.io/kusion/pkg/cmd/generate/run"
	"kusionstack.io/kusion/pkg/cmd/meta"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/spec"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var (
	generateLong = i18n.T(`
		Generate versioned Spec of target Stack. 
	
		The command must be executed in a Stack or by specifying a Stack directory with the -w flag.`)

	generateExample = i18n.T(`
		# Generate spec with working directory
		kusion generate -w /path/to/stack

		# Generate spec with custom workspace
		kusion generate -w /path/to/stack --workspace dev

		# Generate spec with custom backend
		kusion generate -w /path/to/stack --backend oss`)
)

// GenerateFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// GenerateOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type GenerateFlags struct {
	WorkDir   string
	MetaFlags *meta.MetaFlags

	genericiooptions.IOStreams
}

// GenerateOptions defines flags and other configuration parameters for the `generate` command.
type GenerateOptions struct {
	*meta.MetaOptions

	WorkDir string

	SpecStorage spec.Storage
}

// NewGenerateFlags returns a default GenerateFlags
func NewGenerateFlags(streams genericiooptions.IOStreams) *GenerateFlags {
	return &GenerateFlags{
		MetaFlags: meta.NewMetaFlags(),
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o, err := flags.ToOptions()
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())
			return
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for a cli.
func (flags *GenerateFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	flags.MetaFlags.AddFlags(cmd)

	cmd.Flags().StringVarP(&flags.WorkDir, "workdir", "w", flags.WorkDir, i18n.T("The working directory for generate (default is current dir where executed)."))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *GenerateFlags) ToOptions() (*GenerateOptions, error) {
	// If working directory not specified, use current dir where executed
	workDir := flags.WorkDir
	if len(workDir) == 0 {
		workDir, _ = os.Getwd()
	}

	// Convert meta options
	metaOptions, err := flags.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	// Get target spec storage
	specStorage, err := backend.NewSpecStorage(
		*flags.MetaFlags.Backend,
		metaOptions.RefProject.Name,
		metaOptions.RefStack.Name,
		metaOptions.RefWorkspace.Name,
	)
	if err != nil {
		return nil, err
	}

	o := &GenerateOptions{
		WorkDir:     workDir,
		MetaOptions: metaOptions,
		SpecStorage: specStorage,
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
	versionedSpec, err := GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, true)
	if err != nil {
		return err
	}
	return o.SpecStorage.Apply(versionedSpec)
}

// GenerateSpecWithSpinner calls generator to generate versioned Spec. Add a method wrapper for testing purposes.
func GenerateSpecWithSpinner(project *v1.Project, stack *v1.Stack, workspace *v1.Workspace, noStyle bool) (*v1.Spec, error) {
	// Construct generator instance
	defaultGenerator := &generator.DefaultGenerator{
		Project:   project,
		Stack:     stack,
		Workspace: workspace,
		Runner:    &run.KPMRunner{},
	}

	kclPkg, err := api.GetKclPackage(stack.Path)
	if err != nil {
		return nil, err
	}
	defaultGenerator.KclPkg = kclPkg

	var sp *pterm.SpinnerPrinter
	if noStyle {
		fmt.Printf("Generating Spec in the Stack %s...\n", stack.Name)
	} else {
		sp = &pretty.SpinnerT
		sp, _ = sp.Start(fmt.Sprintf("Generating Spec in the Stack %s...", stack.Name))
	}

	// style means color and prompt here. Currently, sp will be nil only when o.NoStyle is true
	style := !noStyle && sp != nil

	versionedSpec, err := defaultGenerator.Generate(stack.Path, nil)
	if err != nil {
		if style {
			sp.Fail()
			return nil, err
		} else {
			return nil, err
		}
	}

	// success
	if style {
		sp.Success()
	} else {
		fmt.Println()
	}

	return versionedSpec, nil
}

func SpecFromFile(filePath string) (*v1.Spec, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// TODO: here we use decoder in yaml.v3 to parse resources because it converts
	// map into map[string]interface{} by default which is inconsistent with yaml.v2.
	// The use of yaml.v2 and yaml.v3 should be unified in the future.
	decoder := yamlv3.NewDecoder(bytes.NewBuffer(b))
	decoder.KnownFields(true)
	i := &v1.Spec{}
	if err = decoder.Decode(i); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse the intent file, please check if the file content is valid")
	}
	return i, nil
}
