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

package destroy

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/liu-hm19/pterm"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/cmd/meta"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/pretty"
	"kusionstack.io/kusion/pkg/util/signal"
	"kusionstack.io/kusion/pkg/util/terminal"
)

var (
	destroyLong = i18n.T(`
		Destroy resources within the stack.

		Please note that the destroy command does NOT perform resource version checks.
		Therefore, if someone submits an update to a resource at the same time you execute a destroy command, 
		their update will be lost along with the rest of the resource.`)

	destroyExample = i18n.T(`
		# Delete resources of current stack
		kusion destroy`)
)

// DeleteFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// DestroyOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type DeleteFlags struct {
	MetaFlags *meta.MetaFlags

	Operator string
	Yes      bool
	Detail   bool
	NoStyle  bool

	UI *terminal.UI

	genericiooptions.IOStreams
}

// DestroyOptions defines flags and other configuration parameters for the `delete` command.
type DestroyOptions struct {
	*meta.MetaOptions

	Yes     bool
	Detail  bool
	NoStyle bool

	UI *terminal.UI

	genericiooptions.IOStreams
}

// NewDeleteFlags returns a default DeleteFlags
func NewDeleteFlags(ui *terminal.UI, streams genericiooptions.IOStreams) *DeleteFlags {
	return &DeleteFlags{
		MetaFlags: meta.NewMetaFlags(),
		UI:        ui,
		IOStreams: streams,
	}
}

// NewCmdDestroy creates the `delete` command.
func NewCmdDestroy(ui *terminal.UI, ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewDeleteFlags(ui, ioStreams)

	cmd := &cobra.Command{
		Use:     "destroy",
		Short:   "Destroy resources within the stack.",
		Long:    templates.LongDesc(destroyLong),
		Example: templates.Examples(destroyExample),
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
func (flags *DeleteFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	flags.MetaFlags.AddFlags(cmd)

	cmd.Flags().BoolVarP(&flags.Yes, "yes", "y", false, i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&flags.Detail, "detail", "d", false, i18n.T("Automatically show preview details after previewing it"))
	cmd.Flags().BoolVarP(&flags.NoStyle, "no-style", "", false, i18n.T("no-style sets to RawOutput mode and disables all of styling"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *DeleteFlags) ToOptions() (*DestroyOptions, error) {
	// Convert meta options
	metaOptions, err := flags.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &DestroyOptions{
		MetaOptions: metaOptions,
		Detail:      flags.Detail,
		Yes:         flags.Yes,
		NoStyle:     flags.NoStyle,
		UI:          flags.UI,
		IOStreams:   flags.IOStreams,
	}

	return o, nil
}

// Validate verifies if DestroyOptions are valid and without conflicts.
func (o *DestroyOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `delete` command.
func (o *DestroyOptions) Run() (err error) {
	// update release to succeeded or failed
	var storage release.Storage
	var rel *apiv1.Release
	releaseCreated := false
	defer func() {
		if !releaseCreated {
			return
		}
		if err != nil {
			rel.Phase = apiv1.ReleasePhaseFailed
			_ = release.UpdateDestroyRelease(storage, rel)
		} else {
			rel.Phase = apiv1.ReleasePhaseSucceeded
			err = release.UpdateDestroyRelease(storage, rel)
		}
	}()

	// only destroy resources we managed
	storage, err = o.Backend.ReleaseStorage(o.RefProject.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}
	rel, err = release.CreateDestroyRelease(storage, o.RefProject.Name, o.RefStack.Name, o.RefWorkspace.Name)
	if err != nil {
		return
	}
	if len(rel.Spec.Resources) == 0 {
		pterm.Println(pterm.Green("No managed resources to destroy"))
		return
	}
	releaseCreated = true

	errCh := make(chan error, 1)
	defer close(errCh)

	// wait for the SIGTERM or SIGINT
	go func() {
		stopCh := signal.SetupSignalHandler()
		<-stopCh
		errCh <- errors.New("receive SIGTERM or SIGINT, exit cmd")
	}()

	// run destroy command
	go func() {
		errCh <- o.run(rel, storage)
	}()

	if err = <-errCh; err != nil {
		rel.Phase = apiv1.ReleasePhaseFailed
		release.UpdateDestroyRelease(storage, rel)
	} else {
		rel.Phase = apiv1.ReleasePhaseSucceeded
		release.UpdateDestroyRelease(storage, rel)
	}

	return err
}

// run executes the delete command after release is created.
func (o *DestroyOptions) run(rel *apiv1.Release, storage release.Storage) (err error) {
	// set no style
	if o.NoStyle {
		pterm.DisableStyling()
	}

	sp := o.UI.SpinnerPrinter
	sp, _ = sp.Start(fmt.Sprintf("Computing destroy changes in the Stack %s...", o.RefStack.Name))

	// compute changes for preview
	changes, err := o.preview(rel.Spec, rel.State, o.RefProject, o.RefStack, storage)
	if err != nil {
		if sp != nil {
			sp.Fail()
		}
		return
	}

	if sp != nil {
		sp.Success()
	}

	// preview
	changes.Summary(os.Stdout, o.NoStyle)

	// detail detection
	if o.Detail {
		changes.OutputDiff("all")
		return nil
	}

	// prompt
	if !o.Yes {
		for {
			var input string
			input, err = prompt(o.UI, rel, storage)
			if err != nil {
				return
			}
			if input == "yes" {
				break
			} else if input == "details" {
				var target string
				target, err = changes.PromptDetails(o.UI)
				if err != nil {
					return
				}
				changes.OutputDiff(target)
			} else {
				fmt.Println("Operation destroy canceled")
				return nil
			}
		}
	}

	// update release phase to destroying
	rel.Phase = apiv1.ReleasePhaseDestroying
	if err = release.UpdateDestroyRelease(storage, rel); err != nil {
		return
	}
	// destroy
	fmt.Println("Start destroying resources......")
	var updatedRel *apiv1.Release
	updatedRel, err = o.destroy(rel, changes, storage)
	if err != nil {
		return err
	}
	rel = updatedRel

	return nil
}

func (o *DestroyOptions) preview(
	planResources *apiv1.Spec,
	priorResources *apiv1.State,
	proj *apiv1.Project,
	stack *apiv1.Stack,
	storage release.Storage,
) (*models.Changes, error) {
	log.Info("Start compute preview changes ...")

	// check and install terraform executable binary for
	// resources with the type of Terraform.
	tfInstaller := terraform.CLIInstaller{
		Intent: planResources,
	}
	if err := tfInstaller.CheckAndInstall(); err != nil {
		return nil, err
	}

	pc := &operation.PreviewOperation{
		Operation: models.Operation{
			OperationType:  models.DestroyPreview,
			Stack:          stack,
			ReleaseStorage: storage,
			ChangeOrder:    &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project: proj,
			Stack:   stack,
		},
		Spec:  planResources,
		State: priorResources,
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return models.NewChanges(proj, stack, rsp.Order), nil
}

func (o *DestroyOptions) destroy(rel *apiv1.Release, changes *models.Changes, storage release.Storage) (*apiv1.Release, error) {
	destroyOpt := &operation.DestroyOperation{
		Operation: models.Operation{
			Stack:          changes.Stack(),
			ReleaseStorage: storage,
			MsgCh:          make(chan models.Message),
		},
	}

	// line summary
	var deleted int

	// progress bar, print dag walk detail
	progressbar, err := o.UI.ProgressbarPrinter.
		WithMaxWidth(0).
		WithTotal(len(changes.StepKeys)).
		WithWriter(o.IOStreams.Out).
		WithRemoveWhenDone().
		Start()
	if err != nil {
		return nil, err
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
			case msg, ok := <-destroyOpt.MsgCh:
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
					pretty.SuccessT.WithWriter(o.IOStreams.Out).Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					deleted++
				case models.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pretty.ErrorT.WithWriter(o.IOStreams.Out).Printf("%s\n", title)
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

	req := &operation.DestroyRequest{
		Request: models.Request{
			Project: changes.Project(),
			Stack:   changes.Stack(),
		},
		Release: rel,
	}
	rsp, status := destroyOpt.Destroy(req)
	if v1.IsErr(status) {
		return nil, fmt.Errorf("destroy failed, status: %v", status)
	}
	updatedRel := rsp.Release

	// wait for msgCh closed
	wg.Wait()
	// print summary
	pterm.Println()
	pterm.Fprintln(o.IOStreams.Out, fmt.Sprintf("Destroy complete! Resources: %d deleted.", deleted))
	return updatedRel, nil
}

func prompt(ui *terminal.UI, rel *apiv1.Release, storage release.Storage) (string, error) {
	options := []string{"yes", "details", "no"}
	input, err := ui.InteractiveSelectPrinter.
		WithFilter(false).
		WithDefaultText(`Do you want to destroy these diffs?`).
		WithOptions(options).
		WithDefaultOption("details").
		// To gracefully exit if interrupted by SIGINT or SIGTERM.
		WithOnInterruptFunc(func() {
			rel.Phase = apiv1.ReleasePhaseFailed
			release.UpdateDestroyRelease(storage, rel)
			os.Exit(1)
		}).
		Show()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return "", err
	}

	return input, nil
}
