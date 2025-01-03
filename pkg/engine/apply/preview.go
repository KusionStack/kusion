package apply

import (
	"fmt"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/apply/options"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/log"
)

// The Preview function calculates the upcoming actions of each resource
// through the execution Kusion Engine, and you can customize the
// runtime of engine and the state storage through `runtime` and
// `storage` parameters.
//
// Example:
//
//	o := newPreviewOptions()
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
	o options.PreviewOptions,
	storage release.Storage,
	planResources *apiv1.Spec,
	priorResources *apiv1.State,
	project *apiv1.Project,
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
			OperationType:  models.ApplyPreview,
			Stack:          stack,
			ReleaseStorage: storage,
			IgnoreFields:   o.GetIgnoreFields(),
			ChangeOrder:    &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project: project,
			Stack:   stack,
		},
		Spec:  planResources,
		State: priorResources,
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return models.NewChanges(project, stack, rsp.Order), nil
}
