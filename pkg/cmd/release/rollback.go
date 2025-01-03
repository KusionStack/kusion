package rel

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/liu-hm19/pterm"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/cmd/apply"
	"kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/util/signal"
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

var errExit = errors.New("receive SIGTERM or SIGINT, exit cmd")

// RollbackFlags reflects the information that CLI is gathering via flags,
// which will be converted into RollbackOptions.
type RollbackFlags struct {
	*apply.ApplyFlags

	Revision    uint64
	Yes         bool
	DryRun      bool
	Watch       bool
	Timeout     int
	PortForward int

	genericiooptions.IOStreams
}

// RollbackOptions defines the configuration parameters for the `kusion release rollback` command.
type RollbackOptions struct {
	*apply.ApplyOptions
	*release.State

	Revision    uint64
	Yes         bool
	DryRun      bool
	Watch       bool
	Timeout     int
	PortForward int

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

	cmd.Flags().BoolVarP(&f.Yes, "yes", "y", false, i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&f.DryRun, "dry-run", "", false, i18n.T("Preview the execution effect (always successful) without actually rollback the changes"))
	cmd.Flags().BoolVarP(&f.Watch, "watch", "", true, i18n.T("After creating/updating/deleting the requested object, watch for changes"))
	cmd.Flags().IntVarP(&f.Timeout, "timeout", "", 0, i18n.T("The timeout duration for kusion release rollback command, measured in second(s)"))
	cmd.Flags().IntVarP(&f.PortForward, "port-forward", "", 0, i18n.T("Forward the specified port from local to service"))
	cmd.Flags().Uint64VarP(&f.Revision, "revision", "", 0, i18n.T("The revision number of the release to rollback to"))
}

// ToOptions converts RollbackFlags to RollbackOptions.
func (f *RollbackFlags) ToOptions() (*RollbackOptions, error) {
	// Convert preview options
	applyOptions, err := f.ApplyFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	rollbackState := release.NewState(applyOptions.RefProject.Name, applyOptions.RefWorkspace.Name)

	o := &RollbackOptions{
		ApplyOptions: applyOptions,
		Yes:          f.Yes,
		DryRun:       f.DryRun,
		Watch:        f.Watch,
		Timeout:      f.Timeout,
		PortForward:  f.PortForward,
		IOStreams:    f.IOStreams,
		State:        rollbackState,
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

// Run executes the `release rollback` command.
func (o *RollbackOptions) Run() (err error) {
	state := o.State
	// update release to succeeded or failed
	defer func() {
		if !state.ReleaseHasStorage {
			return
		}
		if err != nil {
			release.UpdateReleasePhase(state.TargetRel, apiv1.ReleasePhaseFailed, state.RelLock)
			// Join the errors if update apply release failed.
			err = errors.Join([]error{err, release.UpdateApplyRelease(state.ReleaseStorage, state.TargetRel, o.DryRun, state.RelLock)}...)
		} else {
			release.UpdateReleasePhase(state.TargetRel, apiv1.ReleasePhaseSucceeded, state.RelLock)
			err = release.UpdateApplyRelease(state.ReleaseStorage, state.TargetRel, o.DryRun, state.RelLock)
		}
	}()

	// set no style
	if o.NoStyle {
		pterm.DisableStyling()
	}

	// create release
	releaseStorage, err := o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}
	state.ReleaseStorage = releaseStorage
	rel, err := state.NewReleaseByRevision(o.Revision)
	if err != nil {
		return
	}

	if !o.DryRun {
		if err = state.CreateStorageRelease(rel); err != nil {
			return
		}
	}

	// Prepare for the timeout timer and error channel.
	var timer <-chan time.Time
	errCh := make(chan error, 1)
	stopCh := signal.SetupSignalHandler()
	defer close(errCh)

	// Start the main task in a goroutine.
	go func() {
		errCh <- o.run(rel, state)
	}()

	// If timeout is set, initialize the timer.
	if o.Timeout > 0 {
		timer = time.After(time.Second * time.Duration(o.Timeout))
	}

	// Centralized event handling.
	for {
		select {
		case <-stopCh:
			// Handle SIGINT or SIGTERM
			err = errExit
			if state.PortForwarded {
				return nil
			}
			return err

		case err = <-errCh:
			// Handle task completion
			if errors.Is(err, errExit) && state.PortForwarded {
				return nil
			}
			return err

		case <-timer:
			// Handle timeout
			err = fmt.Errorf("failed to execute kusion apply as: timeout for %d seconds", o.Timeout)
			if state.ReleaseHasStorage {
				release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, state.RelLock)
				updateErr := release.UpdateApplyRelease(state.ReleaseStorage, rel, o.DryRun, state.RelLock)
				err = errors.Join([]error{err, updateErr}...)
			}
			return err
		}
	}
}

// run executes the rollback cmd after the release is get.
func (o *RollbackOptions) run(rel *apiv1.Release, state *release.State) (err error) {
	defer func() {
		if !o.ReleaseHasStorage {
			return
		}
		if err != nil {
			release.UpdateReleasePhase(rel, apiv1.ReleasePhaseFailed, o.RelLock)
			err = errors.Join([]error{err, release.UpdateApplyRelease(state.ReleaseStorage, rel, o.DryRun, o.RelLock)}...)
		}
	}()

	release.UpdateReleasePhase(rel, apiv1.ReleasePhasePreviewing, o.RelLock)
	if err = release.UpdateApplyRelease(state.ReleaseStorage, rel, o.DryRun, o.RelLock); err != nil {
		return
	}

	// compute changes for preview
	changes, err := preview.Preview(o.PreviewOptions, state.ReleaseStorage, rel.Spec, rel.State, o.RefProject, o.RefStack)
	if err != nil {
		return
	}

	if allUnChange(changes) {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// summary preview table
	changes.Summary(o.IOStreams.Out, o.NoStyle)

	// detail detection
	if o.Detail && o.All {
		changes.OutputDiff("all")
		if !o.Yes {
			return nil
		}
	}

	// prompt
	if !o.Yes {
		for {
			var input string
			input, err = prompt(o.UI, state)
			if err != nil {
				return err
			}
			if input == "yes" {
				break
			} else if input == "details" {
				var target string
				target, err = changes.PromptDetails(o.UI)
				if err != nil {
					return err
				}
				changes.OutputDiff(target)
			} else {
				fmt.Println("Operation apply canceled")
				return nil
			}
		}
	}

	// update release phase to rollbacking
	release.UpdateReleasePhase(rel, apiv1.ReleasePhaseRollbacking, state.RelLock)
	if err = release.UpdateApplyRelease(state.ReleaseStorage, rel, o.DryRun, state.RelLock); err != nil {
		return
	}

	// Get graph storage directory, create if not exist
	graphStorage, err := o.Backend.GraphStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return err
	}

	// Try to get existing graph, use the graph if exists
	if graphStorage.CheckGraphStorageExistence() {
		state.Gph, err = graphStorage.Get()
		if err != nil {
			return err
		}
		err = graph.ValidateGraph(state.Gph)
		if err != nil {
			return err
		}
		// Put new resources from the generated spec to graph
		state.Gph, err = graph.GenerateGraph(state.TargetRel.Spec.Resources, state.Gph)
	} else {
		// Create a new graph to be used globally if no graph is stored in the storage
		state.Gph = &apiv1.Graph{
			Project:   o.RefProject.Name,
			Workspace: o.RefWorkspace.Name,
		}
		state.Gph, err = graph.GenerateGraph(state.TargetRel.Spec.Resources, state.Gph)
	}
	if err != nil {
		return err
	}

	// start applying
	fmt.Printf("\nStart applying diffs ...\n")

	// NOTE: release should be updated in the process of apply, so as to avoid the problem
	// of being unable to update after being terminated by SIGINT or SIGTERM.
	_, err = apply.Apply(o.ApplyOptions, state.ReleaseStorage, rel, state.Gph, changes)
	if err != nil {
		return
	}

	// if dry run, print the hint
	if o.DryRun {
		fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
		return nil
	}

	if o.PortForward > 0 {
		fmt.Printf("\nStart port-forwarding ...\n")
		o.State.PortForwarded = true
		if err = apply.PortForward(o.ApplyOptions, rel.Spec); err != nil {
			return
		}
	}

	return
}

func allUnChange(changes *models.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != models.UnChanged {
			return false
		}
	}

	return true
}

func prompt(ui *terminal.UI, state *release.State) (string, error) {
	// don`t display yes item when only preview
	options := []string{"yes", "details", "no"}
	input, err := ui.InteractiveSelectPrinter.
		WithFilter(false).
		WithDefaultText(`Do you want to apply these diffs?`).
		WithOptions(options).
		WithDefaultOption("details").
		// To gracefully exit if interrupted by SIGINT or SIGTERM.
		WithOnInterruptFunc(func() {
			release.UpdateReleasePhase(state.TargetRel, apiv1.ReleasePhaseFailed, state.RelLock)
			release.UpdateApplyRelease(state.ReleaseStorage, state.TargetRel, false, state.RelLock)
			os.Exit(1)
		}).
		Show()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return "", err
	}

	return input, nil
}
