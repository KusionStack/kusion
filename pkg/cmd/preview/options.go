package preview

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"

	compilecmd "kusionstack.io/kusion/pkg/cmd/compile"
	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type PreviewOptions struct {
	compilecmd.CompileOptions
	PreviewFlags
	backend.BackendOps
}

type PreviewFlags struct {
	Operator     string
	Detail       bool
	All          bool
	NoStyle      bool
	IgnoreFields []string
}

func NewPreviewOptions() *PreviewOptions {
	return &PreviewOptions{
		CompileOptions: *compilecmd.NewCompileOptions(),
	}
}

func (o *PreviewOptions) Complete(args []string) {
	o.CompileOptions.Complete(args)
}

func (o *PreviewOptions) Validate() error {
	return o.CompileOptions.Validate()
}

func (o *PreviewOptions) Run() error {
	// Set no style
	if o.NoStyle {
		pterm.DisableStyling()
		pterm.EnableColor()
	}

	// Parse project and stack of work directory
	project, stack, err := projectstack.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	// Get compile result
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

	// Compute changes for preview
	changes, err := Preview(o, stateStorage, sp, project, stack)
	if err != nil {
		return err
	}

	if changes.AllUnChange() {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// Summary preview table
	changes.Summary(os.Stdout)

	// Detail detection
	if o.Detail {
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
	}

	return nil
}

// The Preview function calculates the upcoming actions of each resource
// through the execution Kusion Engine, and you can customize the
// runtime of engine and the state storage through `runtime` and
// `storage` parameters.
//
// Example:
//
//	o := NewPreviewOptions()
//	stateStorage := &states.FileSystemState{
//	    Path: filepath.Join(o.WorkDir, states.KusionState)
//	}
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	changes, err := Preview(o, kubernetesRuntime, stateStorage,
//	    planResources, project, stack, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Preview(
	o *PreviewOptions,
	storage states.StateStorage,
	planResources *models.Spec,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	// Validate secret stores
	if !project.SecretStores.IsValid() {
		return nil, fmt.Errorf("no secret store is provided")
	}

	// Construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType: opsmodels.ApplyPreview,
			Stack:         stack,
			StateStorage:  storage,
			IgnoreFields:  o.IgnoreFields,
			ChangeOrder:   &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
			SecretStores:  project.SecretStores,
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	cluster := util.ParseClusterArgument(o.Arguments)
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Tenant:   project.Tenant,
			Project:  project,
			Stack:    stack,
			Operator: o.Operator,
			Spec:     planResources,
			Cluster:  cluster,
		},
	})
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return opsmodels.NewChanges(project, stack, rsp.Order), nil
}
