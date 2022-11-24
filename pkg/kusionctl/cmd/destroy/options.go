package destroy

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/compile"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/types"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeInit "kusionstack.io/kusion/pkg/engine/runtime/init"
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
	backend.BackendOps
}

func NewDestroyOptions() *DestroyOptions {
	return &DestroyOptions{
		CompileOptions: *compilecmd.NewCompileOptions(),
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
	planResources, err := compile.GenerateSpec(&compile.Options{
		WorkDir:     o.WorkDir,
		Filenames:   o.Filenames,
		Settings:    o.Settings,
		Arguments:   o.Arguments,
		Overrides:   o.Overrides,
		DisableNone: o.DisableNone,
		OverrideAST: o.OverrideAST,
	}, stack)
	if err != nil {
		return err
	}

	if planResources == nil || len(planResources.Resources) == 0 {
		pterm.Println("No resources to destroy")
		return nil
	}

	runtimes := runtimeInit.InitRuntime()
	runtime, err := runtimes[planResources.Resources[0].Type]()
	if err != nil {
		return nil
	}

	// Get stateStroage from backend config to manage state
	stateStorage, err := backend.BackendFromConfig(project.Backend, o.BackendOps, o.WorkDir)
	if err != nil {
		return err
	}

	// Compute changes for preview
	changes, err := o.preview(planResources, project, stack, runtime, stateStorage)
	if err != nil {
		return err
	}

	// Preview
	changes.Summary(os.Stdout)

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
	if err := o.destroy(planResources, changes, runtime, stateStorage); err != nil {
		return err
	}
	return nil
}

func (o *DestroyOptions) preview(planResources *models.Spec,
	project *projectstack.Project, stack *projectstack.Stack, runtime runtime.Runtime, stateStorage states.StateStorage,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType: types.DestroyPreview,
			Runtime:       runtime,
			StateStorage:  stateStorage,
			ChangeOrder:   &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Tenant:   project.Tenant,
			Project:  project.Name,
			Operator: o.Operator,
			Stack:    stack.Name,
			Spec:     planResources,
		},
	})
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return opsmodels.NewChanges(project, stack, rsp.Order), nil
}

func (o *DestroyOptions) destroy(planResources *models.Spec, changes *opsmodels.Changes, runtime runtime.Runtime, stateStorage states.StateStorage) error {
	// Build apply operation

	do := &operation.DestroyOperation{
		Operation: opsmodels.Operation{
			Runtime:      runtime,
			StateStorage: stateStorage,
			MsgCh:        make(chan opsmodels.Message),
		},
	}

	// line summary
	var deleted int

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.WithTotal(len(changes.StepKeys)).Start()
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
				case opsmodels.Success, opsmodels.Skip:
					var title string
					if changeStep.Action == types.UnChange {
						title = fmt.Sprintf("%s %s, %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(opsmodels.Skip)),
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
				case opsmodels.Failed:
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
		Request: opsmodels.Request{
			Tenant:   changes.Project().Tenant,
			Project:  changes.Project().Name,
			Operator: o.Operator,
			Stack:    changes.Stack().Name,
			Spec:     planResources,
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
