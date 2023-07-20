package get

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	previewcmd "kusionstack.io/kusion/pkg/cmd/preview"
	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type GetOptions struct {
	previewcmd.PreviewOptions
	GetFlags
}

type GetFlags struct {
	ShowDrift bool
}

func NewGetOptions() *GetOptions {
	return &GetOptions{
		PreviewOptions: *previewcmd.NewPreviewOptions(),
	}
}

func Watch(
	o *GetOptions,
	planResources *models.Spec,
	changes *opsmodels.Changes,
) error {
	toBeGet := planResources.Resources

	// Watch operation
	wo := &operation.WatchOperation{}
	if err := wo.Watch(&operation.WatchRequest{
		Request: opsmodels.Request{
			Project: changes.Project(),
			Stack:   changes.Stack(),
			Spec:    &models.Spec{Resources: toBeGet},
		},
	}); err != nil {
		return err
	}

	fmt.Println("Watch Finish!")
	return nil
}

func (o *GetOptions) Complete(args []string) {
	o.CompileOptions.Complete(args)
}

func (o *GetOptions) Validate() error {
	return o.CompileOptions.Validate()
}

func (o *GetOptions) Run() error {
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

	// generate Spec
	sp, err := spec.GenerateSpecWithSpinner(&generator.Options{
		WorkDir:     o.WorkDir,
		Filenames:   o.Filenames,
		Settings:    o.Settings,
		Arguments:   o.Arguments,
		Overrides:   o.Overrides,
		DisableNone: o.DisableNone,
		OverrideAST: o.OverrideAST,
		NoStyle:     o.NoStyle,
	}, project, stack)
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	if sp == nil || len(sp.Resources) == 0 {
		fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		return nil
	}

	// Get state storage from backend config to manage state
	stateStorage, err := backend.BackendFromConfig(project.Backend, o.BackendOps, o.WorkDir)
	if err != nil {
		return err
	}

	// TODO: --detail/--all flag support
	// // Detail detection
	// if o.Detail && o.All {
	// 	changes.OutputDiff("all")
	// 	return nil
	// }

	if o.ShowDrift {
		// Compute changes for preview
		changes, err := previewcmd.Preview(&o.PreviewOptions, stateStorage, sp, project, stack)
		if err != nil {
			return err
		}

		if allUnChange(changes) {
			fmt.Println("All resources are reconciled. No diff found")
			return nil
		}

		// Summary preview table
		changes.Summary(os.Stdout)

		// Prompt
		for {
			target, err := changes.PromptDetails()
			if err != nil {
				return err
			}
			if target == "" { // Cancel option
				break
			}
			changes.OutputDiff(target)
		}
	} else {
		// TODO: add the `cluster` field in query
		query := &states.StateQuery{
			Tenant:  project.Tenant,
			Stack:   stack.Name,
			Project: project.Name,
		}

		latestState, err := stateStorage.GetLatestState(query)
		if err != nil || latestState == nil {
			// log.Infof("can't find states with query: %v", jsonutil.Marshal2PrettyString(query))
			return fmt.Errorf("can not find State in this stack")
		}

		getResources := latestState.Resources

		if getResources == nil || len(latestState.Resources) == 0 {
			pterm.Println(pterm.Green("No managed resources to get"))
			return nil
		}

		// Compute changes for preview
		spec := &models.Spec{Resources: getResources}
		changes, err := previewcmd.Preview(&o.PreviewOptions, stateStorage, spec, project, stack)
		if err != nil {
			return err
		}

		// Summary preview table
		changes.Summary(os.Stdout)

		fmt.Println("\nStart watching status ...")
		if err := Watch(o, spec, changes); err != nil {
			return err
		}
	}

	return nil
}

func allUnChange(changes *opsmodels.Changes) bool {
	for _, v := range changes.ChangeSteps {
		if v.Action != opsmodels.UnChanged {
			return false
		}
	}

	return true
}
