package destroy

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/compile"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	compilecmd "kusionstack.io/kusion/pkg/kusionctl/cmd/compile"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/signals"
)

type DestroyOptions struct {
	compilecmd.CompileOptions
	Operator string
	Yes      bool
	Detail   bool
}

func NewDestroyOptions() *DestroyOptions {
	return &DestroyOptions{
		CompileOptions: compilecmd.CompileOptions{
			Filenames: []string{},
			Arguments: []string{},
			Settings:  []string{},
			Overrides: []string{},
		},
	}
}

func (o *DestroyOptions) Complete(args []string) {
	o.CompileOptions.Complete(args)
}

func (o *DestroyOptions) Validate() error {
	return o.CompileOptions.Validate()
}

func (o *DestroyOptions) Run() error {
	// listen for interrupts or the SIGTERM signal
	signals.HandleInterrupt()
	// Parse project and stack of work directory
	project, stack, err := projectstack.DetectProjectAndStack(o.CompileOptions.WorkDir)
	if err != nil {
		return err
	}

	// Get compile result
	planResources, sp, err := compile.CompileWithSpinner(o.CompileOptions.WorkDir, o.CompileOptions.Filenames, o.CompileOptions.Settings, o.CompileOptions.Arguments, o.Overrides, stack)
	if err != nil {
		sp.Fail()
		return err
	}
	sp.Success() // Resolve spinner with success message.
	pterm.Println()

	if planResources == nil || len(planResources.Resources) == 0 {
		pterm.Println("No resources to destroy")
		return nil
	}

	// Compute changes for preview
	changes, err := o.preview(planResources, project, stack)
	if err != nil {
		return err
	}

	// Preview
	changes.Summary()

	// Detail detection
	if o.Detail {
		changes.OutputDiff("all")
		return nil
	}
	// Prompt
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
				fmt.Println("Operation destroy canceled")
				return nil
			}
		}
	}

	// Destroy
	fmt.Println("Start destroying resources......")
	if err := o.destroy(planResources, changes); err != nil {
		return err
	}
	return nil
}

func (o *DestroyOptions) preview(planResources *models.Spec,
	project *projectstack.Project, stack *projectstack.Stack,
) (*operation.Changes, error) {
	log.Info("Start compute preview changes ...")

	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
	if err != nil {
		return nil, err
	}

	pc := &operation.PreviewOperation{
		Operation: operation.Operation{
			Runtime:       kubernetesRuntime,
			StateStorage:  &states.FileSystemState{Path: filepath.Join(o.CompileOptions.WorkDir, states.KusionState)},
			ChangeStepMap: map[string]*operation.ChangeStep{},
		},
	}

	log.Info("Start call pc.Preview() ...")

	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: operation.Request{
			Tenant:   project.Tenant,
			Project:  project.Name,
			Operator: o.Operator,
			Stack:    stack.Name,
			Manifest: planResources,
		},
	}, operation.Destroy)
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return operation.NewChanges(project, stack, rsp.ChangeSteps), nil
}

func (o *DestroyOptions) destroy(planResources *models.Spec, changes *operation.Changes) error {
	// Build apply operation
	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
	if err != nil {
		return err
	}

	do := &operation.DestroyOperation{
		Operation: operation.Operation{
			Runtime:       kubernetesRuntime,
			StateStorage:  &states.FileSystemState{Path: filepath.Join(o.CompileOptions.WorkDir, states.KusionState)},
			MsgCh:         make(chan operation.Message),
			ChangeStepMap: map[string]*operation.ChangeStep{},
		},
	}

	// line summary
	var deleted int

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.WithTotal(len(changes.ChangeSteps)).Start()
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
			case msg, ok := <-do.MsgCh:
				if !ok {
					wg.Done()
					return
				}
				changeStep := changes.Get(msg.ResourceID)

				switch msg.OpResult {
				case operation.Success, operation.Skip:
					var title string
					if changeStep.Action == operation.UnChange {
						title = fmt.Sprintf("%s %s, %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(operation.Skip)),
						)
					} else {
						title = fmt.Sprintf("%s %s %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(msg.OpResult)),
						)
					}
					pterm.Success.Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					deleted++
				case operation.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pterm.Error.Printf("%s, %v\n", title, msg.OpErr)
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

	st := do.Destroy(&operation.DestroyRequest{
		Request: operation.Request{
			Tenant:   changes.Project().Tenant,
			Project:  changes.Project().Name,
			Operator: o.Operator,
			Stack:    changes.Stack().Name,
			Manifest: planResources,
		},
	})
	if status.IsErr(st) {
		return fmt.Errorf("destroy failed, status: %v", st)
	}

	// wait for msgCh closed
	wg.Wait()
	// Print summary
	pterm.Println()
	pterm.Printf("Destroy complete! Resources: %d deleted.\n", deleted)
	return nil
}

func prompt() (string, error) {
	prompt := &survey.Select{
		Message: `Do you want to destroy these diffs?`,
		Options: []string{"yes", "details", "no"},
		Default: "details",
	}

	var input string
	err := survey.AskOne(prompt, &input)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}
	return input, nil
}
