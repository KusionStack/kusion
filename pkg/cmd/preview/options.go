package preview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/pkg/errors"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/apis/status"
	"kusionstack.io/kusion/pkg/cmd/build"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/pretty"
)

const jsonOutput = "json"

type Options struct {
	build.Options
	Flags
	backend.BackendOps
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
	o.Options.Complete(args)
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
	// Set no style
	if o.NoStyle || o.Output == jsonOutput {
		pterm.DisableStyling()
		pterm.DisableColor()
	}
	// Parse project and stack of work directory
	project, stack, err := project.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	options := &builders.Options{
		IsKclPkg:  o.IsKclPkg,
		WorkDir:   o.WorkDir,
		Filenames: o.Filenames,
		Settings:  o.Settings,
		Arguments: o.Arguments,
		NoStyle:   o.NoStyle,
	}

	// Generate Intent
	var sp *intent.Intent
	if o.IntentFile != "" {
		sp, err = build.IntentFromFile(o.IntentFile)
	} else if o.Output == jsonOutput {
		sp, err = build.Intent(options, project, stack)
	} else {
		sp, err = build.IntentWithSpinner(options, project, stack)
	}
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if sp == nil || len(sp.Resources) == 0 {
		if o.Output != jsonOutput {
			fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		}
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
	o *Options,
	storage states.StateStorage,
	planResources *intent.Intent,
	project *project.Project,
	stack *stack.Stack,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	// Construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType: opsmodels.ApplyPreview,
			Stack:         stack,
			StateStorage:  storage,
			IgnoreFields:  o.IgnoreFields,
			ChangeOrder:   &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	cluster := o.Arguments["cluster"]
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Tenant:   project.Tenant,
			Project:  project,
			Stack:    stack,
			Operator: o.Operator,
			Intent:   planResources,
			Cluster:  cluster,
		},
	})
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return opsmodels.NewChanges(project, stack, rsp.Order), nil
}
