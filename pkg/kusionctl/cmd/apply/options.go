package apply

import (
	"fmt"
	"io"
	"os"
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
)

type ApplyOptions struct {
	compilecmd.CompileOptions
	Operator    string
	Yes         bool
	Detail      bool
	NoStyle     bool
	DryRun      bool
	OnlyPreview bool
}

func NewApplyOptions() *ApplyOptions {
	return &ApplyOptions{
		CompileOptions: compilecmd.CompileOptions{
			Filenames: []string{},
			Arguments: []string{},
			Settings:  []string{},
			Overrides: []string{},
		},
	}
}

func (o *ApplyOptions) Complete(args []string) {
	o.CompileOptions.Complete(args)
}

func (o *ApplyOptions) Validate() error {
	return o.CompileOptions.Validate()
}

func (o *ApplyOptions) Run() error {
	// Set no style
	if o.NoStyle {
		pterm.DisableStyling()
		pterm.EnableColor()
	}

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

	// Compute changes for preview
	stateStorage := &states.FileSystemState{Path: filepath.Join(o.WorkDir, states.KusionState)}
	changes, err := Preview(o, stateStorage, planResources, project, stack, os.Stdout)
	if err != nil {
		return err
	}

	if allUnChange(changes) {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// Summary preview table
	changes.Summary()

	// Detail detection
	if o.Detail && !o.Yes {
		changes.OutputDiff("all")
		return nil
	}

	// Prompt
	if !o.Yes {
		for {
			input, err := prompt(o.OnlyPreview)
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
			}

			fmt.Println("Operation apply canceled")
			return nil
		}
	}

	if !o.OnlyPreview {
		fmt.Println("Start applying diffs ...")
		if err := Apply(o, stateStorage, planResources, changes, os.Stdout); err != nil {
			return err
		}

		// If dry run, print the hint
		if o.DryRun {
			fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
		}
	}

	return nil
}

// The preview function will preview for all resources changes
func Preview(
	o *ApplyOptions,
	storage states.StateStorage,
	planResources *models.Spec,
	project *projectstack.Project,
	stack *projectstack.Stack,
	out io.Writer,
) (*operation.Changes, error) {
	log.Info("Start compute preview changes ...")

	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
	if err != nil {
		return nil, err
	}

	pc := &operation.PreviewOperation{
		Operation: operation.Operation{
			Runtime:      kubernetesRuntime,
			StateStorage: storage,
			Order:        &operation.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*operation.ChangeStep{}},
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
	}, operation.Apply)
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return operation.NewChanges(project, stack, rsp.Order), nil
}

// The apply function will apply the resources changes,
// and will save the state to specified storage.
func Apply(
	o *ApplyOptions,
	storage states.StateStorage,
	planResources *models.Spec,
	changes *operation.Changes,
	out io.Writer,
) error {
	// Build apply operation
	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
	if err != nil {
		return err
	}

	ac := &operation.ApplyOperation{
		Operation: operation.Operation{
			Runtime:      kubernetesRuntime,
			StateStorage: storage,
			MsgCh:        make(chan operation.Message),
		},
	}

	// Line summary
	var ls lineSummary

	// Progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.WithTotal(len(changes.StepKeys)).WithWriter(out).Start()
	if err != nil {
		return err
	}
	// Wait msgCh close
	var wg sync.WaitGroup
	// Receive msg and print detail
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
					pterm.Success.WithWriter(out).Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					ls.Count(changeStep.Action)
				case operation.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pterm.Error.WithWriter(out).Printf("%s, %v\n", title, msg.OpErr)
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
			ac.MsgCh <- operation.Message{
				ResourceID: r.ResourceKey(),
				OpResult:   operation.Success,
				OpErr:      nil,
			}
		}
		close(ac.MsgCh)
	} else {
		_, st := ac.Apply(&operation.ApplyRequest{
			Request: operation.Request{
				Tenant:   changes.Project().Tenant,
				Project:  changes.Project().Name,
				Operator: o.Operator,
				Stack:    changes.Stack().Name,
				Manifest: planResources,
			},
		})
		if status.IsErr(st) {
			return fmt.Errorf("apply failed, status: %v", st)
		}
	}

	// Wait for msgCh closed
	wg.Wait()
	// Print summary
	pterm.Fprintln(out)
	pterm.Fprintln(out, fmt.Sprintf("Apply complete! Resources: %d created, %d updated, %d deleted.", ls.created, ls.updated, ls.deleted))
	return nil
}

type lineSummary struct {
	created, updated, deleted int
}

func (ls *lineSummary) Count(op operation.ActionType) {
	switch op {
	case operation.Create:
		ls.created++
	case operation.Update:
		ls.updated++
	case operation.Delete:
		ls.deleted++
	}
}

func allUnChange(changes *operation.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != operation.UnChange {
			return false
		}
	}

	return true
}

func prompt(onlyPreview bool) (string, error) {
	// don`t display yes item when only preview
	options := []string{"details", "no"}
	if !onlyPreview {
		options = append([]string{"yes"}, options...)
	}

	prompt := &survey.Select{
		Message: `Do you want to apply these diffs?`,
		Options: options,
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
