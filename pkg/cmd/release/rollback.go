package rel

import (
	"errors"
	"sync"

	applystate "kusionstack.io/kusion/pkg/engine/apply/state"

	"kusionstack.io/kusion/pkg/cmd/apply"
	applyaction "kusionstack.io/kusion/pkg/engine/apply"
	"kusionstack.io/kusion/pkg/util/terminal"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/util/i18n"

	cmdutil "kusionstack.io/kusion/pkg/cmd/util"

	"github.com/spf13/cobra"
)

var (
	rollbackShort = i18n.T("Rollback to a specific release of the current or specified stack")

	rollbackLong = i18n.T(`
	Rollback to a specific release of the current or specified stack.
	
	This command reverts the current project in the current or a specified workspace 
	to the state of a specified release.
	`)

	rollbackExample = i18n.T(`
	# Rollback to the latest release of the current project in the current workspace
	kusion release rollback

	# Rollback with specified work directory
	kusion release rollback  -w /path/to/workdir

	# Rollback to a specific release of the current project in the current workspace
	kusion release rollback --revision=1

	# Skip interactive approval of preview details before rollback
	kusion release rollback --yes

	# Rollback without output style and color
	kusion release rollback --no-style=true
		
	# Rollback without watching the resource changes and waiting for reconciliation
	kusion release rollback --watch=false

	# Rollback with the specified timeout duration for kusion apply command, measured in second(s)
	kusion release rollback --timeout=120

	# Rollback with localhost port forwarding
	kusion release rollback --port-forward=8080
	`)
)

// RollbackFlags reflects the information that CLI is gathering via flags,
// which will be converted into RollbackOptions.
type RollbackFlags struct {
	*apply.ApplyFlags

	Revision uint64

	genericiooptions.IOStreams
}

// RollbackOptions defines the configuration parameters for the `kusion release rollback` command.
type RollbackOptions struct {
	*apply.ApplyOptions

	Revision uint64

	genericiooptions.IOStreams
}

// NewRollbackFlags returns a default RollbackFlags.
func NewRollbackFlags(ui *terminal.UI, streams genericiooptions.IOStreams) *RollbackFlags {
	return &RollbackFlags{
		ApplyFlags: apply.NewApplyFlags(ui, streams),
		IOStreams:  streams,
	}
}

// NewCmdRollback creates the `kusion release rollback` command.
func NewCmdRollback(ui *terminal.UI, streams genericiooptions.IOStreams) *cobra.Command {
	flags := NewRollbackFlags(ui, streams)

	cmd := &cobra.Command{
		Use:     "rollback",
		Short:   rollbackShort,
		Long:    templates.LongDesc(rollbackLong),
		Example: templates.Examples(rollbackExample),
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

// AddFlags adds flags for a RollbackOptions struct to the specified command.
func (f *RollbackFlags) AddFlags(cmd *cobra.Command) {
	f.PreviewFlags.AddFlags(cmd)

	cmd.Flags().Uint64VarP(&f.Revision, "revision", "", 0, i18n.T("The revision number of the release to rollback to"))
}

// ToOptions converts RollbackFlags to RollbackOptions.
func (f *RollbackFlags) ToOptions() (*RollbackOptions, error) {
	// Convert preview options
	applyOptions, err := f.ApplyFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &RollbackOptions{
		ApplyOptions: applyOptions,
		Revision:     f.Revision,
		IOStreams:    f.IOStreams,
	}

	return o, nil
}

// Validate checks the provided options for the `kusion release rollback` command.
func (o *RollbackOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	if o.PortForward < 0 || o.PortForward > 65535 {
		return cmdutil.UsageErrorf(cmd, "Invalid port number to forward: %d, must be between 1 and 65535", o.PortForward)
	}
	return nil
}

func (o *RollbackOptions) prepareRollback() (state *applystate.State, err error) {
	defer cmdutil.RecoverErr(&err)

	// init apply state
	state = &applystate.State{
		Metadata: &applystate.Metadata{
			Project:   o.RefProject.Name,
			Workspace: o.RefWorkspace.Name,
			Stack:     o.RefStack.Name,
		},
		RelLock:          &sync.Mutex{},
		PortForward:      o.PortForward,
		DryRun:           o.DryRun,
		Watch:            o.Watch,
		CallbackRevision: o.Revision,
		Ls:               &applystate.LineSummary{},
	}

	// create release
	state.ReleaseStorage, err = o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}

	state.TargetRel, err = release.NewRollbackRelease(state.ReleaseStorage, o.RefProject.Name, o.RefStack.Name, o.RefWorkspace.Name, o.Revision)
	if err != nil {
		return
	}

	if !o.DryRun {
		if err = state.CreateStorageRelease(state.TargetRel); err != nil {
			return
		}
	}
	return
}

// Run executes the `release rollback` command.
func (o *RollbackOptions) Run() (err error) {
	// prepare apply
	applyState, err := o.prepareRollback()

	if err != nil && applyState != nil {
		updateErr := applyState.UpdateReleasePhaseFailed()
		if updateErr != nil {
			err = errors.Join(err, updateErr)
		}
	}

	if err != nil {
		return
	}

	// apply action
	err = applyaction.Apply(o.ApplyOptions, applyState)
	return
}
