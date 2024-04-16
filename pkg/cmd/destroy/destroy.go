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
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
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
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/pretty"
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
// DeleteOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type DeleteFlags struct {
	MetaFlags *meta.MetaFlags

	Operator string
	Yes      bool
	Detail   bool

	genericiooptions.IOStreams
}

// DeleteOptions defines flags and other configuration parameters for the `delete` command.
type DeleteOptions struct {
	*meta.MetaOptions

	Operator string
	Yes      bool
	Detail   bool

	genericiooptions.IOStreams
}

// NewDeleteFlags returns a default DeleteFlags
func NewDeleteFlags(streams genericiooptions.IOStreams) *DeleteFlags {
	return &DeleteFlags{
		MetaFlags: meta.NewMetaFlags(),
		IOStreams: streams,
	}
}

// NewCmdDestroy creates the `delete` command.
func NewCmdDestroy(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewDeleteFlags(ioStreams)

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

	cmd.Flags().StringVarP(&flags.Operator, "operator", "", flags.Operator, i18n.T("Specify the operator"))
	cmd.Flags().BoolVarP(&flags.Yes, "yes", "y", false, i18n.T("Automatically approve and perform the update after previewing it"))
	cmd.Flags().BoolVarP(&flags.Detail, "detail", "d", false, i18n.T("Automatically show preview details after previewing it"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *DeleteFlags) ToOptions() (*DeleteOptions, error) {
	// Convert meta options
	metaOptions, err := flags.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &DeleteOptions{
		MetaOptions: metaOptions,
		Operator:    flags.Operator,
		Detail:      flags.Detail,
		Yes:         flags.Yes,
		IOStreams:   flags.IOStreams,
	}

	return o, nil
}

// Validate verifies if DeleteOptions are valid and without conflicts.
func (o *DeleteOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	return nil
}

// Run executes the `delete` command.
func (o *DeleteOptions) Run() error {
	// only destroy resources we managed
	storage := o.StorageBackend.StateStorage(o.RefProject.Name, o.RefWorkspace.Name)
	priorState, err := storage.Get()
	if err != nil || priorState == nil {
		return fmt.Errorf("can not find DeprecatedState in this stack")
	}
	destroyResources := priorState.Resources

	if destroyResources == nil || len(priorState.Resources) == 0 {
		pterm.Println(pterm.Green("No managed resources to destroy"))
		return nil
	}

	// compute changes for preview
	i := &apiv1.Spec{Resources: destroyResources}
	changes, err := o.preview(i, o.RefProject, o.RefStack, storage)
	if err != nil {
		return err
	}

	// preview
	changes.Summary(os.Stdout, false)

	// detail detection
	if o.Detail {
		changes.OutputDiff("all")
		return nil
	}
	// prompt
	if !o.Yes {
		for {
			var input string
			input, err = prompt()
			if err != nil {
				return err
			}

			if input == "yes" {
				break
			} else if input == "details" {
				var target string
				target, err = changes.PromptDetails()
				if err != nil {
					return err
				}
				changes.OutputDiff(target)
			} else {
				fmt.Println("Operation destroy canceled")
				return nil
			}
		}
	}

	// destroy
	fmt.Println("Start destroying resources......")
	if err = o.destroy(i, changes, storage); err != nil {
		return err
	}
	return nil
}

func (o *DeleteOptions) preview(
	planResources *apiv1.Spec,
	proj *apiv1.Project,
	stack *apiv1.Stack,
	stateStorage state.Storage,
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
			OperationType: models.DestroyPreview,
			Stack:         stack,
			StateStorage:  stateStorage,
			ChangeOrder:   &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project:  proj,
			Stack:    stack,
			Operator: o.Operator,
			Intent:   planResources,
		},
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return models.NewChanges(proj, stack, rsp.Order), nil
}

func (o *DeleteOptions) destroy(planResources *apiv1.Spec, changes *models.Changes, stateStorage state.Storage) error {
	destroyOpt := &operation.DestroyOperation{
		Operation: models.Operation{
			Stack:        changes.Stack(),
			StateStorage: stateStorage,
			MsgCh:        make(chan models.Message),
		},
	}

	// line summary
	var deleted int

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.WithMaxWidth(0).WithTotal(len(changes.StepKeys)).Start()
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
					pretty.SuccessT.Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					deleted++
				case models.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pretty.ErrorT.Printf("%s\n", title)
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

	st := destroyOpt.Destroy(&operation.DestroyRequest{
		Request: models.Request{
			Project:  changes.Project(),
			Stack:    changes.Stack(),
			Operator: o.Operator,
			Intent:   planResources,
		},
	})
	if v1.IsErr(st) {
		return fmt.Errorf("destroy failed, status: %v", st)
	}

	// wait for msgCh closed
	wg.Wait()
	// print summary
	pterm.Println()
	pterm.Printf("Destroy complete! Resources: %d deleted.\n", deleted)
	return nil
}

func prompt() (string, error) {
	p := &survey.Select{
		Message: `Do you want to destroy these diffs?`,
		Options: []string{"yes", "details", "no"},
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
