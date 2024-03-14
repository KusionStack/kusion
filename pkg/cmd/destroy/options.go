package destroy

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/build"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/signals"
)

type Options struct {
	build.Options
	Operator string
	Yes      bool
	Detail   bool
}

func NewDestroyOptions() *Options {
	return &Options{
		Options: *build.NewBuildOptions(),
	}
}

func (o *Options) Complete(args []string) {
	_ = o.Options.Complete(args)
}

func (o *Options) Validate() error {
	if err := o.Options.Validate(); err != nil {
		return err
	}
	return nil
}

func (o *Options) Run() error {
	// listen for interrupts or the SIGTERM signal
	signals.HandleInterrupt()
	// Parse project and stack of work directory
	proj, stack, err := project.DetectProjectAndStack(o.Options.WorkDir)
	if err != nil {
		return err
	}

	// new state storage
	ws := "default"                   // fixme: use default workspace for tmp
	bk, err := backend.NewBackend("") // fixme: use current backend for tmp
	if err != nil {
		return err
	}
	storage := bk.StateStorage(proj.Name, stack.Name, ws)

	// only destroy resources we managed
	priorState, err := storage.Get()
	if err != nil || priorState == nil {
		log.Infof("can't find state with project: %s, stack: %s, workspace: %s", proj.Name, stack.Name, ws)
		return fmt.Errorf("can not find State in this stack")
	}
	destroyResources := priorState.Resources

	if destroyResources == nil || len(priorState.Resources) == 0 {
		pterm.Println(pterm.Green("No managed resources to destroy"))
		return nil
	}

	// Compute changes for preview
	i := &apiv1.Intent{Resources: destroyResources}
	changes, err := o.preview(i, proj, stack, ws, storage)
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

	// Destroy
	fmt.Println("Start destroying resources......")
	if err = o.destroy(i, changes, storage); err != nil {
		return err
	}
	return nil
}

func (o *Options) preview(
	planResources *apiv1.Intent,
	proj *apiv1.Project,
	stack *apiv1.Stack,
	ws string,
	stateStorage state.Storage,
) (*models.Changes, error) {
	log.Info("Start compute preview changes ...")

	// Check and install terraform executable binary for
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
			Project:   proj,
			Stack:     stack,
			Workspace: ws,
			Operator:  o.Operator,
			Intent:    planResources,
		},
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return models.NewChanges(proj, stack, ws, rsp.Order), nil
}

func (o *Options) destroy(planResources *apiv1.Intent, changes *models.Changes, stateStorage state.Storage) error {
	do := &operation.DestroyOperation{
		Operation: models.Operation{
			Stack:        changes.Stack(),
			StateStorage: stateStorage,
			MsgCh:        make(chan models.Message),
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
					pterm.Success.Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					deleted++
				case models.Failed:
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
		Request: models.Request{
			Project:   changes.Project(),
			Stack:     changes.Stack(),
			Workspace: changes.Workspace(),
			Operator:  o.Operator,
			Intent:    planResources,
		},
	})
	if v1.IsErr(st) {
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
