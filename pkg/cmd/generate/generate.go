package generate

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/generate/generator"
	"kusionstack.io/kusion/pkg/cmd/generate/run"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/spec"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/i18n"
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
	Backend   string
	Workspace string

	genericiooptions.IOStreams
}

// GenerateOptions defines flags and other configuration parameters for the `generate` command.
type GenerateOptions struct {
	WorkDir string

	Project   *v1.Project
	Stack     *v1.Stack
	Workspace *v1.Workspace

	SpecStorage spec.Storage
	Generator   generator.Generator
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
	cmd.Flags().StringVarP(&flags.WorkDir, "workdir", "w", flags.WorkDir, i18n.T("The working directory for generate (default is current dir where executed)."))
	cmd.Flags().StringVarP(&flags.Backend, "backend", "", flags.Backend, i18n.T("The backend to use, supports 'local', 'oss' and 's3'."))
	cmd.Flags().StringVarP(&flags.Workspace, "workspace", "", flags.Workspace, i18n.T("The name of target workspace to operate in."))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *GenerateFlags) ToOptions() (*GenerateOptions, error) {
	// If working directory not specified, use current dir where executed
	workDir := flags.WorkDir
	if len(workDir) == 0 {
		workDir, _ = os.Getwd()
	}

	// Parse project and currentStack of work directory
	currentProject, currentStack, err := project.DetectProjectAndStack(workDir)
	if err != nil {
		return nil, err
	}

	// Get current workspace from backend
	workspaceStorage, err := backend.NewWorkspaceStorage(flags.Backend)
	if err != nil {
		return nil, err
	}
	currentWorkspace, err := workspaceStorage.Get(flags.Workspace)
	if err != nil {
		return nil, err
	}

	// Get target spec storage
	specStorage, err := backend.NewSpecStorage(flags.Backend, currentProject.Name, currentStack.Name, flags.Workspace)
	if err != nil {
		return nil, err
	}

	kclPkg, err := api.GetKclPackage(workDir)
	if err != nil {
		return nil, err
	}

	// Construct generator instance
	defaultGenerator := &generator.DefaultGenerator{
		Project:   currentProject,
		Stack:     currentStack,
		Workspace: currentWorkspace,
		Runner:    &run.KPMRunner{},
		KclPkg:    kclPkg,
	}

	o := &GenerateOptions{
		WorkDir:     workDir,
		SpecStorage: specStorage,
		Generator:   defaultGenerator,
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
	return o.SpecStorage.Apply(intent)
}
