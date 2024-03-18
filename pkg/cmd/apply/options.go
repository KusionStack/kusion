package apply

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/build"
	"kusionstack.io/kusion/pkg/cmd/preview"
	engineapi "kusionstack.io/kusion/pkg/engine/api"
	"kusionstack.io/kusion/pkg/engine/api/builders"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/pretty"
)

// Options defines flags for the `apply` command
type Options struct {
	preview.Options
	Flag
}

type Flag struct {
	Yes    bool
	DryRun bool
	Watch  bool
}

// NewApplyOptions returns a new ApplyOptions instance
func NewApplyOptions() *Options {
	return &Options{
		Options: *preview.NewPreviewOptions(),
	}
}

func (o *Options) Complete(args []string) {
	o.Options.Complete(args)
}

func (o *Options) Validate() error {
	return o.Options.Validate()
}

func (o *Options) Run() error {
	// set no style
	if o.NoStyle {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	options := &builders.Options{
		IsKclPkg:  o.IsKclPkg,
		WorkDir:   o.WorkDir,
		Filenames: o.Filenames,
		Settings:  o.Settings,
		Arguments: o.Arguments,
		NoStyle:   o.NoStyle,
	}

	// parse project and stack of work directory
	proj, stack, err := project.DetectProjectAndStack(o.Options.WorkDir)
	if err != nil {
		return err
	}

	// get workspace configurations
	bk, err := backend.NewBackend(o.Backend)
	if err != nil {
		return err
	}
	wsStorage, err := bk.WorkspaceStorage()
	if err != nil {
		return err
	}
	ws, err := wsStorage.Get(o.Workspace)
	if err != nil {
		return err
	}

	// generate Intent
	var sp *apiv1.Intent
	if o.IntentFile != "" {
		sp, err = build.IntentFromFile(o.IntentFile)
	} else {
		sp, err = build.IntentWithSpinner(options, proj, stack, ws)
	}
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	if sp == nil || len(sp.Resources) == 0 {
		fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		return nil
	}

	// new state storage
	storage := bk.StateStorage(proj.Name, stack.Name, ws.Name)

	// Compute changes for preview
	// Construct sdk option
	previewOptions := &engineapi.APIOptions{
		Operator:     o.Operator,
		Cluster:      o.Arguments["cluster"],
		IgnoreFields: o.IgnoreFields,
	}
	changes, err := engineapi.Preview(previewOptions, storage, sp, proj, stack, ws)
	if err != nil {
		return err
	}

	if allUnChange(changes) {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// Summary preview table
	changes.Summary(os.Stdout, false)

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
			input, err := prompt()
			if err != nil {
				return err
			}
			if input == "yes" {
				break
			} else if input == "details" {
				target, err := changes.PromptDetails()
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

	fmt.Println("Start applying diffs ...")
	if err = Apply(o, storage, sp, changes, os.Stdout); err != nil {
		return err
	}

	// if dry run, print the hint
	if o.DryRun {
		fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
		return nil
	}

	if o.Watch {
		fmt.Println("\nStart watching changes ...")
		if err = Watch(o, sp, changes); err != nil {
			return err
		}
	}

	return nil
}

// The Apply function will apply the resources changes
// through the execution Kusion Engine, and will save
// the state to specified storage.
//
// You can customize the runtime of engine and the state
// storage through `runtime` and `storage` parameters.
//
// Example:
//
//	o := NewApplyOptions()
//	stateStorage := &states.FileSystemState{
//	    Path: filepath.Join(o.WorkDir, states.KusionState)
//	}
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	err = Apply(o, kubernetesRuntime, stateStorage, planResources, changes, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Apply(
	o *Options,
	storage state.Storage,
	planResources *apiv1.Intent,
	changes *models.Changes,
	out io.Writer,
) error {
	// construct the apply operation
	ac := &operation.ApplyOperation{
		Operation: models.Operation{
			Stack:        changes.Stack(),
			StateStorage: storage,
			MsgCh:        make(chan models.Message),
			IgnoreFields: o.IgnoreFields,
		},
	}

	// line summary
	var ls lineSummary

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.
		WithMaxWidth(0). // Set to 0, the terminal width will be used
		WithTotal(len(changes.StepKeys)).
		WithWriter(out).
		Start()
	if err != nil {
		return err
	}
	// wait msgCh close
	var wg sync.WaitGroup
	// receive msg and print detail
	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Errorf("failed to receive msg and print detail as %v", p)
			}
		}()
		wg.Add(1)

		for {
			select {
			case msg, ok := <-ac.MsgCh:
				if !ok {
					wg.Done()
					return
				}
				changeStep := changes.Get(msg.ResourceID)

				switch msg.OpResult {
				case models.Success, models.Skip:
					var title string
					if changeStep.Action == models.UnChanged {
						title = fmt.Sprintf("%s %s, %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(models.Skip)),
						)
					} else {
						title = fmt.Sprintf("%s %s %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(msg.OpResult)),
						)
					}
					pterm.Success.WithWriter(out).Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					ls.Count(changeStep.Action)
				case models.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pterm.Error.WithWriter(out).Printf("%s\n", title)
				default:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.Ing(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					progressbar.UpdateTitle(title)
				}
			}
		}
	}()

	if o.DryRun {
		for _, r := range planResources.Resources {
			ac.MsgCh <- models.Message{
				ResourceID: r.ResourceKey(),
				OpResult:   models.Success,
				OpErr:      nil,
			}
		}
		close(ac.MsgCh)
	} else {
		// parse cluster in arguments
		_, st := ac.Apply(&operation.ApplyRequest{
			Request: models.Request{
				Project:  changes.Project(),
				Stack:    changes.Stack(),
				Operator: o.Operator,
				Intent:   planResources,
			},
		})
		if v1.IsErr(st) {
			return fmt.Errorf("apply failed, status:\n%v", st)
		}
	}

	// wait for msgCh closed
	wg.Wait()
	// print summary
	pterm.Fprintln(out, fmt.Sprintf("Apply complete! Resources: %d created, %d updated, %d deleted.", ls.created, ls.updated, ls.deleted))
	return nil
}

// Watch function will observe the changes of each resource
// by the execution engine.
//
// Example:
//
//	o := NewApplyOptions()
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	Watch(o, kubernetesRuntime, planResources, changes, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Watch(
	o *Options,
	planResources *apiv1.Intent,
	changes *models.Changes,
) error {
	if o.DryRun {
		fmt.Println("NOTE: Watch doesn't work in DryRun mode")
		return nil
	}

	// filter out unchanged resources
	toBeWatched := apiv1.Resources{}
	for _, res := range planResources.Resources {
		if changes.ChangeOrder.ChangeSteps[res.ResourceKey()].Action != models.UnChanged {
			toBeWatched = append(toBeWatched, res)
		}
	}

	// watch operation
	wo := &operation.WatchOperation{}
	if err := wo.Watch(&operation.WatchRequest{
		Request: models.Request{
			Project: changes.Project(),
			Stack:   changes.Stack(),
			Intent:  &apiv1.Intent{Resources: toBeWatched},
		},
	}); err != nil {
		return err
	}

	fmt.Println("Watch Finish! All resources have been reconciled.")
	return nil
}

type lineSummary struct {
	created, updated, deleted int
}

func (ls *lineSummary) Count(op models.ActionType) {
	switch op {
	case models.Create:
		ls.created++
	case models.Update:
		ls.updated++
	case models.Delete:
		ls.deleted++
	}
}

func allUnChange(changes *models.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != models.UnChanged {
			return false
		}
	}

	return true
}

func prompt() (string, error) {
	// don`t display yes item when only preview
	options := []string{"yes", "details", "no"}

	p := &survey.Select{
		Message: `Do you want to apply these diffs?`,
		Options: options,
		Default: "details",
	}

	var input string
	err := survey.AskOne(p, &input)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}
	return input, nil
}
