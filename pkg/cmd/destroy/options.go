package destroy

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/status"
	"kusionstack.io/kusion/pkg/cmd/build"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/project"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/signals"
)

type Options struct {
	build.Options
	Operator string
	Yes      bool
	Detail   bool
	backend.BackendOptions
}

func NewDestroyOptions() *Options {
	return &Options{
		Options: *build.NewBuildOptions(),
	}
}

func (o *Options) Complete(args []string) {
	o.Options.Complete(args)
}

func (o *Options) Validate() error {
	if err := o.Options.Validate(); err != nil {
		return err
	}
	if !o.BackendOptions.IsEmpty() {
		if err := o.BackendOptions.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) Run() error {
	// listen for interrupts or the SIGTERM signal
	signals.HandleInterrupt()
	// Parse project and stack of work directory
	project, stack, err := project.DetectProjectAndStack(o.Options.WorkDir)
	if err != nil {
		return err
	}

	// Get state storage from cli backend options, environment variables, workspace backend configs
	stateStorage, err := backend.NewStateStorage(stack, &o.BackendOptions)
	if err != nil {
		return err
	}

	// only destroy resources we managed
	// todo add the `cluster` field in query
	query := &states.StateQuery{
		Tenant:  "",
		Stack:   stack.Name,
		Project: project.Name,
	}
	latestState, err := stateStorage.GetLatestState(query)
	if err != nil || latestState == nil {
		log.Infof("can't find states with query: %v", jsonutil.Marshal2PrettyString(query))
		return fmt.Errorf("can not find State in this stack")
	}
	destroyResources := latestState.Resources

	if destroyResources == nil || len(latestState.Resources) == 0 {
		pterm.Println(pterm.Green("No managed resources to destroy"))
		return nil
	}

	// Compute changes for preview
	i := &intent.Intent{Resources: destroyResources}
	changes, err := o.preview(i, project, stack, stateStorage)
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
	if err := o.destroy(i, changes, stateStorage); err != nil {
		return err
	}
	return nil
}

func (o *Options) preview(
	planResources *intent.Intent, project *apiv1.Project,
	stack *apiv1.Stack, stateStorage states.StateStorage,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType: opsmodels.DestroyPreview,
			Stack:         stack,
			StateStorage:  stateStorage,
			ChangeOrder:   &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Tenant:   "",
			Project:  project,
			Operator: o.Operator,
			Stack:    stack,
			Intent:   planResources,
		},
	})
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return opsmodels.NewChanges(project, stack, rsp.Order), nil
}

func (o *Options) destroy(planResources *intent.Intent, changes *opsmodels.Changes, stateStorage states.StateStorage) error {
	do := &operation.DestroyOperation{
		Operation: opsmodels.Operation{
			Stack:        changes.Stack(),
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
					if changeStep.Action == opsmodels.UnChanged {
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
					pterm.Error.Printf("%s\n", title)
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
			Tenant:   "",
			Project:  changes.Project(),
			Operator: o.Operator,
			Stack:    changes.Stack(),
			Intent:   planResources,
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
