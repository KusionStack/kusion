// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apply

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/cmd/preview"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/util/terminal"
)

var (
	applyLong = i18n.T(`
		Apply a series of resource changes within the stack.
	
		Create, update or delete resources according to the operational intent within a stack.
		By default, Kusion will generate an execution preview and prompt for your approval before performing any actions.
		You can review the preview details and make a decision to proceed with the actions or abort them.`)

	applyExample = i18n.T(`
		# Apply with specified work directory
		kusion apply -w /path/to/workdir

		# Apply with specified arguments
		kusion apply -D name=test -D age=18
	
		# Skip interactive approval of preview details before applying
		kusion apply --yes
		
		# Apply without output style and color
		kusion apply --no-style=true
		
		# Apply with localhost port forwarding
		kusion apply --port-forward=8080`)
)

// ApplyFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// ApplyOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type ApplyFlags struct {
	*preview.PreviewFlags

	Yes         bool
	DryRun      bool
	Watch       bool
	PortForward int

	genericiooptions.IOStreams
}

// ApplyOptions defines flags and other configuration parameters for the `apply` command.
type ApplyOptions struct {
	*preview.PreviewOptions

	Yes         bool
	DryRun      bool
	Watch       bool
	PortForward int

	genericiooptions.IOStreams
}

// NewApplyFlags returns a default ApplyFlags
func NewApplyFlags(ui *terminal.UI, streams genericiooptions.IOStreams) *ApplyFlags {
	return &ApplyFlags{
		PreviewFlags: preview.NewPreviewFlags(ui, streams),
		IOStreams:    streams,
	}
}

// NewCmdApply creates the `apply` command.
func NewCmdApply(ui *terminal.UI, ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewApplyFlags(ui, ioStreams)

	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply the operational intent of various resources to multiple runtimes",
		Long:    templates.LongDesc(applyLong),
		Example: templates.Examples(applyExample),
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
func (f *ApplyFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	f.PreviewFlags.AddFlags(cmd)

	cmd.Flags().BoolVarP(&f.Yes, "yes", "y", false, i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&f.DryRun, "dry-run", "", false, i18n.T("Preview the execution effect (always successful) without actually applying the changes"))
	cmd.Flags().BoolVarP(&f.Watch, "watch", "", false, i18n.T("After creating/updating/deleting the requested object, watch for changes"))
	cmd.Flags().IntVarP(&f.PortForward, "port-forward", "", 0, i18n.T("Forward the specified port from local to service"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (f *ApplyFlags) ToOptions() (*ApplyOptions, error) {
	// Convert preview options
	previewOptions, err := f.PreviewFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &ApplyOptions{
		PreviewOptions: previewOptions,
		Yes:            f.Yes,
		DryRun:         f.DryRun,
		Watch:          f.Watch,
		PortForward:    f.PortForward,
		IOStreams:      f.IOStreams,
	}

	return o, nil
}

// Validate verifies if ApplyOptions are valid and without conflicts.
func (o *ApplyOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	if o.PortForward < 0 || o.PortForward > 65535 {
		return cmdutil.UsageErrorf(cmd, "Invalid port number to forward: %d, must be between 1 and 65535", o.PortForward)
	}

	return nil
}

// Run executes the `apply` command.
func (o *ApplyOptions) Run() error {
	// set no style
	if o.NoStyle {
		pterm.DisableStyling()
	}

	// build parameters
	parameters := make(map[string]string)
	for _, value := range o.PreviewOptions.Values {
		parts := strings.SplitN(value, "=", 2)
		parameters[parts[0]] = parts[1]
	}

	// Generate Spec
	spec, err := generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, nil, o.UI, o.NoStyle)
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	if spec == nil || len(spec.Resources) == 0 {
		fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		return nil
	}

	// compute changes for preview
	storage := o.StorageBackend.StateStorage(o.RefProject.Name, o.RefWorkspace.Name)
	changes, err := preview.Preview(o.PreviewOptions, storage, spec, o.RefProject, o.RefStack)
	if err != nil {
		return err
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
			input, err := prompt(o.UI)
			if err != nil {
				return err
			}
			if input == "yes" {
				break
			} else if input == "details" {
				target, err := changes.PromptDetails(o.UI)
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
	if err = Apply(o, storage, spec, changes, o.IOStreams.Out); err != nil {
		return err
	}

	// if dry run, print the hint
	if o.DryRun {
		fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
		return nil
	}

	if o.Watch {
		fmt.Println("\nStart watching changes ...")
		if err = Watch(o, spec, changes); err != nil {
			return err
		}
	}

	if o.PortForward > 0 {
		fmt.Printf("\nStart port-forwarding ...\n")
		if err = PortForward(o, spec); err != nil {
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
	o *ApplyOptions,
	storage state.Storage,
	planResources *apiv1.Spec,
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
	progressbar, err := o.UI.ProgressbarPrinter.
		WithMaxWidth(0). // Set to 0, the terminal width will be used
		WithTotal(len(changes.StepKeys)).
		WithWriter(out).
		WithRemoveWhenDone().
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
					pretty.SuccessT.WithWriter(out).Printfln(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					ls.Count(changeStep.Action)
				case models.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pretty.ErrorT.WithWriter(out).Printf("%s\n", title)
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
	pterm.Println()
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
	o *ApplyOptions,
	planResources *apiv1.Spec,
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
			Intent:  &apiv1.Spec{Resources: toBeWatched},
		},
	}); err != nil {
		return err
	}

	fmt.Println("Watch Finish! All resources have been reconciled.")
	return nil
}

// PortForward function will forward the specified port from local to the project Kubernetes Service.
//
// Example:
//
// o := NewApplyOptions()
// spec, err := generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, nil, o.NoStyle)
//
//	if err != nil {
//		 return err
//	}
//
// err = PortForward(o, spec)
//
//	if err != nil {
//	  return err
//	}
func PortForward(
	o *ApplyOptions,
	spec *apiv1.Spec,
) error {
	if o.DryRun {
		fmt.Println("NOTE: Portforward doesn't work in DryRun mode")
		return nil
	}

	// portforward operation
	wo := &operation.PortForwardOperation{}
	if err := wo.PortForward(&operation.PortForwardRequest{
		Request: models.Request{
			Intent: spec,
		},
		Port: o.PortForward,
	}); err != nil {
		return err
	}

	fmt.Println("Portforward has been completed!")
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

func prompt(ui *terminal.UI) (string, error) {
	// don`t display yes item when only preview
	options := []string{"yes", "details", "no"}
	input, err := ui.InteractiveSelectPrinter.
		WithFilter(false).
		WithDefaultText(`Do you want to apply these diffs?`).
		WithOptions(options).
		WithDefaultOption("details").
		Show()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return "", err
	}

	return input, nil
}
