package apply

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/compile"
	"kusionstack.io/kusion/pkg/engine/backend"
	_ "kusionstack.io/kusion/pkg/engine/backend/init"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/operation/types"
	"kusionstack.io/kusion/pkg/engine/runtime"
	runtimeInit "kusionstack.io/kusion/pkg/engine/runtime/init"
	"kusionstack.io/kusion/pkg/engine/states"
	previewcmd "kusionstack.io/kusion/pkg/kusionctl/cmd/preview"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

// ApplyOptions defines flags for the `apply` command
type ApplyOptions struct {
	previewcmd.PreviewOptions
	ApplyFlag
}

type ApplyFlag struct {
	Yes         bool
	NoStyle     bool
	DryRun      bool
	OnlyPreview bool
}

// NewApplyOptions returns a new ApplyOptions instance
func NewApplyOptions() *ApplyOptions {
	return &ApplyOptions{
		PreviewOptions: *previewcmd.NewPreviewOptions(),
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

	// Get state storage from backend config to manage state
	stateStorage, err := backend.BackendFromConfig(project.Backend, o.BackendOps, o.WorkDir)
	if err != nil {
		return err
	}

	// Compute changes for preview
	runtimes := runtimeInit.InitRuntime()
	r, err := runtimes[planResources.Resources[0].Type]()
	if err != nil {
		return err
	}

	changes, err := previewcmd.Preview(&o.PreviewOptions, r, stateStorage, planResources, project, stack, os.Stdout)
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
			} else {
				fmt.Println("Operation apply canceled")
				return nil
			}
		}
	}

	if !o.OnlyPreview {
		fmt.Println("Start applying diffs ...")
		if err := Apply(o, r, stateStorage, planResources, changes, os.Stdout); err != nil {
			return err
		}

		// If dry run, print the hint
		if o.DryRun {
			fmt.Printf("\nNOTE: Currently running in the --dry-run mode, the above configuration does not really take effect\n")
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
	runtime runtime.Runtime,
	storage states.StateStorage,
	planResources *models.Spec,
	changes *opsmodels.Changes,
	out io.Writer,
) error {
	// Construct the apply operation
	ac := &operation.ApplyOperation{
		Operation: opsmodels.Operation{
			Runtime:      runtime,
			StateStorage: storage,
			MsgCh:        make(chan opsmodels.Message),
		},
	}

	// Line summary
	var ls lineSummary

	// Progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.
		WithMaxWidth(0). // Set to 0, the terminal width will be used
		WithTotal(len(changes.StepKeys)).
		WithWriter(out).
		Start()
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
					pterm.Success.WithWriter(out).Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					ls.Count(changeStep.Action)
				case opsmodels.Failed:
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
			ac.MsgCh <- opsmodels.Message{
				ResourceID: r.ResourceKey(),
				OpResult:   opsmodels.Success,
				OpErr:      nil,
			}
		}
		close(ac.MsgCh)
	} else {
		cluster := planResources.ParseCluster()
		_, st := ac.Apply(&operation.ApplyRequest{
			Request: opsmodels.Request{
				Tenant:   changes.Project().Tenant,
				Project:  changes.Project().Name,
				Stack:    changes.Stack().Name,
				Cluster:  cluster,
				Operator: o.Operator,
				Spec:     planResources,
			},
		})
		if status.IsErr(st) {
			return fmt.Errorf("apply failed, status:\n%v", st)
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

func (ls *lineSummary) Count(op types.ActionType) {
	switch op {
	case types.Create:
		ls.created++
	case types.Update:
		ls.updated++
	case types.Delete:
		ls.deleted++
	}
}

func allUnChange(changes *opsmodels.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != types.UnChange {
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
