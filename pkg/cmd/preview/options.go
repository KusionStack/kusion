package preview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pterm/pterm"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/build"
	"kusionstack.io/kusion/pkg/cmd/generate"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/pretty"
)

const jsonOutput = "json"

type Options struct {
	build.Options
	Flags
}

type Flags struct {
	Operator     string
	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	IntentFile   string
	IgnoreFields []string
}

func NewPreviewOptions() *Options {
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
	if o.Output != "" && o.Output != jsonOutput {
		return errors.New("invalid output type, supported types: json")
	}
	if err := o.ValidateIntentFile(); err != nil {
		return err
	}
	return nil
}

func (o *Options) ValidateIntentFile() error {
	if o.IntentFile == "" {
		return nil
	}

	// calculate the absolute path of the intentFile
	var absSF string
	if o.WorkDir == "" {
		absSF, _ = filepath.Abs(o.IntentFile)
	} else if filepath.IsAbs(o.IntentFile) {
		absSF = o.IntentFile
	} else {
		absSF = filepath.Join(o.WorkDir, o.IntentFile)
	}

	fi, err := os.Stat(absSF)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("intent file not exist")
		}
		return err
	}

	if fi.IsDir() || !fi.Mode().IsRegular() {
		return fmt.Errorf("intent file must be a regular file")
	}

	// calculate the relative path between absWD and absSF,
	// if absSF is not located in the directory or subdirectory specified by absWD,
	// an error will be returned
	absWD, _ := filepath.Abs(o.WorkDir)
	rel, err := filepath.Rel(absWD, absSF)
	if err != nil {
		return err
	}
	if rel[:3] == ".."+string(filepath.Separator) {
		return fmt.Errorf("the intent file must be located in the working directory or its subdirectories")
	}

	// set the intent file to the absolute path for further processing
	o.IntentFile = absSF
	return nil
}

func (o *Options) Run() error {
	// set no style
	if o.NoStyle || o.Output == jsonOutput {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	// Parse project and currentStack of work directory
	currentProject, currentStack, err := project.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	// Get current workspace from backend
	bk, err := backend.NewBackend(o.Backend)
	if err != nil {
		return err
	}
	wsStorage, err := bk.WorkspaceStorage()
	if err != nil {
		return err
	}
	currentWorkspace, err := wsStorage.Get(o.Workspace)
	if err != nil {
		return err
	}

	// Generate Spec
	var spec *apiv1.Spec
	if len(o.IntentFile) != 0 {
		spec, err = generate.SpecFromFile(o.IntentFile)
	} else {
		spec, err = generate.GenerateSpecWithSpinner(currentProject, currentStack, currentWorkspace, true)
	}
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if spec == nil || len(spec.Resources) == 0 {
		if o.Output != jsonOutput {
			fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		}
		return nil
	}

	// compute changes for preview
	storage := bk.StateStorage(currentProject.Name, currentStack.Name, currentWorkspace.Name)
	changes, err := Preview(o, storage, spec, currentProject, currentStack)
	if err != nil {
		return err
	}

	if o.Output == jsonOutput {
		var previewChanges []byte
		previewChanges, err = json.Marshal(changes)
		if err != nil {
			return fmt.Errorf("json marshal preview changes failed as %w", err)
		}
		fmt.Println(string(previewChanges))
		return nil
	}

	if changes.AllUnChange() {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// summary preview table
	changes.Summary(os.Stdout)

	// detail detection
	if o.Detail {
		for {
			var target string
			target, err = changes.PromptDetails()
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
	o *Options,
	storage state.Storage,
	planResources *apiv1.Spec,
	proj *apiv1.Project,
	stack *apiv1.Stack,
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

	// construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: models.Operation{
			OperationType: models.ApplyPreview,
			Stack:         stack,
			StateStorage:  storage,
			IgnoreFields:  o.IgnoreFields,
			ChangeOrder:   &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project:  proj,
			Stack:    stack,
			Operator: o.Operator,
			Intent:   planResources,
		},
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return models.NewChanges(proj, stack, rsp.Order), nil
}
