package apply

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/liu-hm19/pterm"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/apply/options"
	applystate "kusionstack.io/kusion/pkg/engine/apply/state"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/signal"
	"kusionstack.io/kusion/pkg/util/terminal"
)

var errExit = errors.New("receive SIGTERM or SIGINT, exit cmd")

// The Apply function will apply the resources changes through the execution kusion engine.
// You can customize the runtime of engine and the release releaseStorage through `runtime` and `releaseStorage` parameters.
func Apply(
	o options.ApplyOptions,
	state *applystate.State,
) (err error) {
	// update release to succeeded or failed
	defer func() {
		if err != nil {
			errUpdate := state.UpdateReleasePhaseFailed()
			if errUpdate != nil {
				// Join the errors if update apply release failed.
				err = errors.Join(err, errUpdate)
			}
			log.Error(err)
		} else {
			err = state.UpdateReleasePhaseSucceeded()
			if err != nil {
				log.Error(err)
			}
		}
		state.ExitClear()
	}()
	defer cmdutil.RecoverErr(&err)

	// set no style
	if o.GetNoStyle() {
		pterm.DisableStyling()
	}
	// Prepare for the timeout timer and error channel.
	var timer <-chan time.Time
	stopCh := signal.SetupSignalHandler()

	// Start the main task in a goroutine.
	taskResult := make(chan error)
	go task(o, state, taskResult)

	// If timeout is set, initialize the timer.
	if o.GetTimeout() > 0 {
		timer = time.After(time.Second * time.Duration(o.GetTimeout()))
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

		case err = <-taskResult:
			// Handle task completion
			if errors.Is(err, errExit) && state.PortForwarded {
				return nil
			}
			return err

		case <-timer:
			// Handle timeout
			err = fmt.Errorf("failed to execute kusion apply as: timeout for %d seconds", o.GetTimeout())
			return err
		}
	}
}

func task(o options.ApplyOptions, state *applystate.State, taskResult chan<- error) {
	var err error
	defer func() {
		if err != nil {
			err = WrappedErr(err, "Apply task err")
		}
		taskResult <- err
		close(taskResult)
	}()

	defer cmdutil.RecoverErr(&err)

	// start preview
	if err = state.UpdateReleasePhasePreviewing(); err != nil {
		return
	}

	var changes *models.Changes

	// compute changes for preview
	changes, err = Preview(o, state.ReleaseStorage, state.TargetRel.Spec, state.TargetRel.State, o.GetRefProject(), o.GetRefStack())
	if err != nil {
		return
	}

	if allUnChange(changes) {
		fmt.Println("All resources are reconciled. No diff found")
		return
	}

	// summary preview table
	changes.Summary(state.IO.Out, o.GetNoStyle())

	// detail detection
	if o.GetDetail() && o.GetAll() {
		changes.OutputDiff("all")
		if !o.GetYes() {
			return
		}
	}

	// prompt
	if !o.GetYes() {
		for {
			var input string
			input, err = prompt(o.GetUI(), state)
			if err != nil {
				return
			}
			if input == "yes" {
				break
			} else if input == "details" {
				var target string
				target, err = changes.PromptDetails(o.GetUI())
				if err != nil {
					return
				}
				changes.OutputDiff(target)
			} else {
				fmt.Println("Operation apply canceled")
				return
			}
		}
	}

	// update release phase to applying
	err = state.UpdateReleasePhaseApplying()
	if err != nil {
		return
	}

	// Get graph storage directory, create if not exist
	state.GraphStorage, err = o.GetBackend().GraphStorage(o.GetRefProject().Name, o.GetRefWorkspace().Name)
	if err != nil {
		return
	}

	// Try to get existing graph, use the graph if exists
	if state.GraphStorage.CheckGraphStorageExistence() {
		state.Gph, err = state.GraphStorage.Get()
		if err != nil {
			return
		}
		err = graph.ValidateGraph(state.Gph)
		if err != nil {
			return
		}
		// Put new resources from the generated spec to graph
		state.Gph, err = graph.GenerateGraph(state.TargetRel.Spec.Resources, state.Gph)
	} else {
		// Create a new graph to be used globally if no graph is stored in the storage
		state.Gph = &apiv1.Graph{
			Project:   o.GetRefProject().Name,
			Workspace: o.GetRefWorkspace().Name,
		}
		state.Gph, err = graph.GenerateGraph(state.TargetRel.Spec.Resources, state.Gph)
	}
	if err != nil {
		return
	}

	// start applying
	fmt.Printf("\nStart applying diffs ...\n")

	err = apply(o, state, changes)
	if err != nil {
		return
	}

	// if dry run, print the hint
	if o.GetDryRun() {
		fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
		return
	}

	if state.PortForward > 0 {
		fmt.Printf("\nStart port-forwarding ...\n")
		state.PortForwarded = true
		if err = PortForward(state); err != nil {
			return
		}
	}
}

func apply(o options.ApplyOptions, state *applystate.State, changes *models.Changes) (err error) {
	// Update the release before exit.
	defer func() {
		var finishErr error
		// Update graph and write to storage if not dry run.
		if !state.DryRun {
			// Use resources in the state to get resource Cloud ID.
			for _, resource := range state.TargetRel.State.Resources {
				var info *graph.ResourceInfo
				// Get information of each of the resources
				info, finishErr = graph.GetResourceInfo(&resource)
				if finishErr != nil {
					err = errors.Join(err, finishErr)
					return
				}
				// Update information of each of the resources.
				graphResource := graph.FindGraphResourceByID(state.Gph.Resources, resource.ID)
				if graphResource != nil {
					graphResource.CloudResourceID = info.CloudResourceID
					graphResource.Type = info.ResourceType
					graphResource.Name = info.ResourceName
				}
			}

			// Update graph if exists, otherwise create a new graph file.
			if state.GraphStorage.CheckGraphStorageExistence() {
				// No need to store resource index
				graph.RemoveResourceIndex(state.Gph)
				finishErr = state.GraphStorage.Update(state.Gph)
				if finishErr != nil {
					err = errors.Join(err, finishErr)
					return
				}
			} else {
				graph.RemoveResourceIndex(state.Gph)
				finishErr = state.GraphStorage.Create(state.Gph)
				if finishErr != nil {
					err = errors.Join(err, finishErr)
					return
				}
			}
		}
	}()

	defer cmdutil.RecoverErr(&err)

	// construct the apply operation
	ac := &operation.ApplyOperation{
		Operation: models.Operation{
			Stack:          changes.Stack(),
			ReleaseStorage: state.ReleaseStorage,
			MsgCh:          make(chan models.Message),
			IgnoreFields:   o.GetIgnoreFields(),
		},
	}

	// Init a watch channel with a sufficient buffer when it is necessary to perform watching.
	isWatch := state.Watch && !state.DryRun
	if isWatch {
		ac.WatchCh = make(chan string, 100)
	}

	// Get the multi printer from UI option.
	multi := o.GetUI().MultiPrinter
	// Max length of resource ID for progressbar width.
	maxLen := 0

	// Prepare the writer to print the operation progress and results.
	changesWriterMap := make(map[string]*pterm.SpinnerPrinter)
	for _, key := range changes.Values() {
		// Get the maximum length of the resource ID.
		if len(key.ID) > maxLen {
			maxLen = len(key.ID)
		}
		// Init a spinner printer for the resource to print the apply status.
		changesWriterMap[key.ID], err = o.GetUI().SpinnerPrinter.
			WithWriter(multi.NewWriter()).
			Start(fmt.Sprintf("Pending %s", pterm.Bold.Sprint(key.ID)))
		if err != nil {
			return fmt.Errorf("failed to init change step spinner printer: %v", err)
		}
	}

	// Init a writer for progressbar.
	pbWriter := multi.NewWriter()
	// progress bar, print dag walk detail
	progressbar, err := o.GetUI().ProgressbarPrinter.
		WithTotal(len(changes.StepKeys)).
		WithWriter(pbWriter).
		WithRemoveWhenDone().
		WithShowCount(false).
		WithMaxWidth(maxLen + 32).
		Start()
	if err != nil {
		return err
	}

	// The writer below is for operation error printing.
	errWriter := multi.NewWriter()

	multi.WithUpdateDelay(time.Millisecond * 100)
	multi.Start()
	defer multi.Stop()

	// apply result
	applyResult := make(chan error)
	// receive msg and print detail
	go PrintApplyDetails(state, ac.MsgCh, applyResult, changesWriterMap, progressbar, changes)

	var watchResult chan error
	// Apply while watching the resources.
	if isWatch {
		watchResult = make(chan error)
		go Watch(state, watchResult, ac.WatchCh, multi, changesWriterMap, changes)
	}

	var updatedRel *apiv1.Release
	if state.DryRun {
		for _, r := range state.TargetRel.Spec.Resources {
			ac.MsgCh <- models.Message{
				ResourceID: r.ResourceKey(),
				OpResult:   models.Success,
				OpErr:      nil,
			}
		}
		close(ac.MsgCh)
	} else {
		// parse cluster in arguments
		rsp, st := ac.Apply(&operation.ApplyRequest{
			Request: models.Request{
				Project: changes.Project(),
				Stack:   changes.Stack(),
			},
			Release: state.TargetRel,
			Graph:   state.Gph,
		})
		if v1.IsErr(st) {
			errWriter.(*bytes.Buffer).Reset()
			err = fmt.Errorf("apply failed, status:\n%v", st)
			return err
		}
		// Update the release with that in the apply response if not dryrun.
		updatedRel = rsp.Release
		if updatedRel != nil {
			*state.TargetRel = *updatedRel
		}
		state.Gph = rsp.Graph
	}

	// wait for apply result ( msgCh closed )
	err = <-applyResult
	if err != nil {
		return
	}

	// Wait for watchWg closed if need to perform watching.
	if isWatch {
		shouldBreak := false
		for !shouldBreak {
			select {
			case watchErr := <-watchResult:
				if watchErr != nil {
					err = watchErr
					return
				}
				shouldBreak = true
			default:
				continue
			}
		}
	}

	// print summary
	pterm.Fprintln(pbWriter, fmt.Sprintf("\nApply complete! Resources: %d created, %d updated, %d deleted.", state.Ls.GetCreated(), state.Ls.GetUpdated(), state.Ls.GetDeleted()))
	return nil
}

func allUnChange(changes *models.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != models.UnChanged {
			return false
		}
	}

	return true
}

func prompt(ui *terminal.UI, state *applystate.State) (string, error) {
	// don`t display yes item when only preview
	ops := []string{"yes", "details", "no"}
	input, err := ui.InteractiveSelectPrinter.
		WithFilter(false).
		WithDefaultText(`Do you want to apply these diffs?`).
		WithOptions(ops).
		WithDefaultOption("details").
		// To gracefully exit if interrupted by SIGINT or SIGTERM.
		WithOnInterruptFunc(func() {
			state.InterruptFunc()
			os.Exit(1)
		}).
		Show()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return "", err
	}

	return input, nil
}
